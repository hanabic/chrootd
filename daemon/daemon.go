package main

import (
	"context"
	"errors"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/consul"
	"github.com/docker/libkv/store/zookeeper"
	"github.com/rs/zerolog"
	etcdv3 "github.com/smallnest/libkv-etcdv3-store"
	"github.com/smallnest/rpcx/log"
	"github.com/smallnest/rpcx/server"
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/cntr"
	"github.com/xhebox/chrootd/utils"
	"github.com/xhebox/libkv-bolt"
)

func init() {
	boltdb.Register()
	etcdv3.Register()
	consul.Register()
	zookeeper.Register()
}

var (
	ErrDaemonStart = errors.New("daemon start")
)

type User struct {
	Logger zerolog.Logger

	ServiceAddr         string
	ServicePath         string
	ServiceReadTimeout  time.Duration
	ServiceWriteTimeout time.Duration
	ServiceRootless bool
	ServiceSecure   bool

	AttachAddr   string
	AttachLimits int

	ConfPath    string
	RunPath     string
	PidFileName string
	LogFileName string
}

func main() {
	u := &User{
		Logger: zerolog.New(os.Stdout).With().Timestamp().Logger(),
	}

	app := &cli.App{
		Usage:                  "chrootd daemon program",
		EnableBashCompletion:   true,
		UseShortOptionHandling: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"c"},
				Value:       "/etc/chrootd/conf",
				EnvVars:     []string{"CHROOTD_CONFIG"},
				Usage:       "load toml config from `FILE`",
				Destination: &u.ConfPath,
			},
			&cli.BoolFlag{
				Name:  "daemon",
				Value: false,
				Usage: "start in background",
			},
			&cli.PathFlag{
				Name:        "pid",
				Value:       "chrootd.pid",
				EnvVars:     []string{"CHROOTD_PIDFILE"},
				Usage:       "daemon `PIDFILE` path",
				Destination: &u.PidFileName,
			},
			&cli.PathFlag{
				Name:        "run",
				Value:       "/var/lib/chrootd",
				EnvVars:     []string{"CHROOTD_RUNPATH"},
				Usage:       "daemon `RUNPATH` for states/persistence",
				Destination: &u.RunPath,
			},
			&cli.PathFlag{
				Name:        "log",
				Value:       "chrootd.log",
				EnvVars:     []string{"CHROOTD_LOGFILE"},
				Usage:       "daemon `LOGFILE` path",
				Destination: &u.LogFileName,
			},
			&cli.StringFlag{
				Name:  "loglevel",
				Value: "info",
				Usage: "set `loglevel`: debug, info, warn, error",
			},
			&cli.StringSliceFlag{
				Name:  "registry",
				Usage: "libkv store `addr`, used for service and container resolution",
			},
			&cli.StringFlag{
				Name:  "registry_backend",
				Usage: "specify the `backend` used by registry, boltdb/zk/consul/etcdv3",
			},
			&cli.StringFlag{
				Name:  "registry_bucket",
				Usage: "if registry is backened by bolt, a `bucket` name is needed",
				Value: "chrootd",
			},
			&cli.StringFlag{
				Name:        "service_path",
				Usage:       "`service` path used by rpcx",
				Value:       "cntr",
				Destination: &u.ServicePath,
			},
			&cli.StringFlag{
				Name:        "service_addr",
				Usage:       "`address` to listen container service",
				Value:       "tcp@:9090",
				EnvVars:     []string{"CHROOTD_ADDR"},
				Destination: &u.ServiceAddr,
			},
			&cli.DurationFlag{
				Name:        "service_readtimeout",
				Usage:       "server read `timeout`",
				Value:       3 * time.Second,
				EnvVars:     []string{"CHROOTD_READTIMEOUT"},
				Destination: &u.ServiceReadTimeout,
			},
			&cli.DurationFlag{
				Name:        "service_writetimeout",
				Usage:       "server write `timeout`",
				Value:       3 * time.Second,
				EnvVars:     []string{"CHROOTD_WRITETIMEOUT"},
				Destination: &u.ServiceWriteTimeout,
			},
			&cli.BoolFlag{
				Name:        "service_rootless",
				Usage:       "service will run in rootless mode",
				Value:       true,
				Destination: &u.ServiceRootless,
			},
			&cli.BoolFlag{
				Name:        "service_secure",
				Usage:       "container will have a safe/default configuration, KNOW WHAT YOU ARE DOING WHEN DISABLING IT",
				Value:       true,
				Destination: &u.ServiceSecure,
			},
			&cli.IntFlag{
				Name:        "attach_limits",
				Value:       64,
				Usage:       "maximum attachable proccess limits",
				Destination: &u.AttachLimits,
			},
			&cli.StringFlag{
				Name:        "attach_addr",
				Usage:       "`address` for process attach",
				Value:       "tcp@:9091",
				Destination: &u.AttachAddr,
			},
		},
		Before: utils.NewTomlFlagLoader("config"),
		Action: func(c *cli.Context) error {
			user := c.Context.Value("_data").(*User)

			if c.Bool("daemon") {
				return ErrDaemonStart
			}

			if !filepath.IsAbs(user.RunPath) {
				return errors.New("run should be absolute")
			}

			if err := os.MkdirAll(user.RunPath, 0755); err != nil {
				return err
			}

			switch c.String("loglevel") {
			case "debug":
				user.Logger = user.Logger.Level(zerolog.DebugLevel)
			case "info":
				user.Logger = user.Logger.Level(zerolog.InfoLevel)
			case "warn":
				user.Logger = user.Logger.Level(zerolog.WarnLevel)
			case "error":
				user.Logger = user.Logger.Level(zerolog.ErrorLevel)
			}

			user.Logger.Log().Msgf("daemon started, logleve - %s", user.Logger.GetLevel())

			log.SetLogger(utils.NewRpcxLogger(user.Logger))

			var registry cntr.Registry

			if backend := c.String("registry_backend"); len(backend) > 0 {
				store, err := libkv.NewStore(
					store.Backend(backend),
					c.StringSlice("registry"),
					&store.Config{Bucket: c.String("registry_bucket")},
				)
				if err != nil {
					return err
				}
				defer store.Close()

				registry = cntr.NewStoreRegistry(store)
			}

			states, err := libkv.NewStore(store.BOLTDB, []string{filepath.Join(user.RunPath, "states")}, &store.Config{Bucket: "states"})
			if err != nil {
				return err
			}
			defer states.Close()

			cntrServer, err := cntr.NewServer(filepath.Join(user.RunPath, "cntrs"),
				user.ServiceAddr,
				user.AttachAddr,
				states,
				registry, func(s *cntr.Server) error {
					s.ProcLimits = user.AttachLimits
					s.Rootless = user.ServiceRootless
					s.Context = c.Context
					s.Secure = user.ServiceSecure
					return nil
				})
			if err != nil {
				return err
			}
			defer cntrServer.Close()

			srv := server.NewServer(
				server.WithReadTimeout(user.ServiceReadTimeout),
				server.WithWriteTimeout(user.ServiceWriteTimeout),
			)

			err = cntrServer.Register(srv, user.ServicePath)
			if err != nil {
				return err
			}

			user.Logger.Log().Msgf("container service server started at %s", user.ServiceAddr)
			user.Logger.Log().Msgf("attach server started at %s", user.AttachAddr)

			go func() {
				h := make(chan os.Signal, 1)
				signal.Notify(h, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
				s := <-h
				switch s {
				case syscall.SIGKILL:
					srv.Close()
				default:
					srv.Shutdown(context.Background())
				}
			}()

			go func() {
				lis := net.ListenConfig{}

				ua := utils.NewAddrFromString(user.AttachAddr)
				listener, err := lis.Listen(c.Context, ua.Network(), ua.Addr())
				if err != nil {
					return
				}
				defer listener.Close()

				for {
					conn, err := listener.Accept()
					if err != nil {
						return
					}

					go cntrServer.ServeAttach(c.Context, conn)
				}
			}()

			ua := utils.NewAddrFromString(user.ServiceAddr)
			return srv.Serve(ua.Network(), ua.Addr())
		},
	}

	ctx := context.WithValue(context.Background(), "_data", u)

	err := app.RunContext(ctx, os.Args)
	if err == ErrDaemonStart {
		skip := false
		args := make([]string, len(os.Args)-1)
		for _, v := range os.Args[1:] {
			switch {
			case v == "--daemon":
				skip = true
			case strings.HasPrefix(v, "--daemon=") || v == "-d" || skip:
				skip = false
			default:
				args = append(args, v)
			}
		}
		cmd := exec.Command(os.Args[0], args...)
		cmd.Start()
	} else if err != nil {
		u.Logger.Fatal().Msg(err.Error())
	}
}
