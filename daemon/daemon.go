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
	"github.com/docker/libkv/store/boltdb"
	"github.com/docker/libkv/store/consul"
	"github.com/docker/libkv/store/etcd"
	"github.com/docker/libkv/store/zookeeper"
	"github.com/rs/zerolog"
	"github.com/smallnest/rpcx/log"
	"github.com/smallnest/rpcx/server"
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/cntr"
	"github.com/xhebox/chrootd/utils"
)

func init() {
	boltdb.Register()
	etcd.Register()
	consul.Register()
	zookeeper.Register()
}

var (
	ErrDaemonStart = errors.New("daemon start")
)

type User struct {
	Logger      zerolog.Logger
	Addr        string
	Timeout     time.Duration
	ConfPath    string
	ServicePath string
	RunPath     string
	PidFileName string
	LogFileName string
	ProcLimits  int
	Rootless    bool
}

func main() {
	u := &User{
		Logger: zerolog.New(os.Stdout).With().Timestamp().Logger(),
	}

	app := &cli.App{
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
			&cli.StringFlag{
				Name:        "srvpath",
				Usage:       "service path",
				Value:       "cntr",
				Destination: &u.ServicePath,
			},
			&cli.StringFlag{
				Name:        "addr",
				Usage:       "for container service listening",
				Value:       "tcp@:9090",
				EnvVars:     []string{"CHROOTD_ADDR"},
				Destination: &u.Addr,
			},
			&cli.StringSliceFlag{
				Name:  "registry",
				Usage: "libkv store addr, used for service and container resolution",
			},
			&cli.StringFlag{
				Name:  "regback",
				Usage: "specify the backend used by registry",
			},
			&cli.StringFlag{
				Name:  "registry_bucket",
				Usage: "if registry is backened by bolt, a bucket name is needed",
				Value: "chrootd",
			},
			&cli.DurationFlag{
				Name:        "timeout",
				Usage:       "server connection timeout",
				Value:       10 * time.Second,
				EnvVars:     []string{"CHROOTD_TIMEOUT"},
				Destination: &u.Timeout,
			},
			&cli.PathFlag{
				Name:        "pid",
				Value:       "chrootd.pid",
				EnvVars:     []string{"CHROOTD_PIDFILE"},
				Usage:       "daemon pid path",
				Destination: &u.PidFileName,
			},
			&cli.PathFlag{
				Name:        "runpath",
				Value:       "/var/lib/chrootd",
				EnvVars:     []string{"CHROOTD_RUNPATH"},
				Usage:       "daemon run path",
				Destination: &u.RunPath,
			},
			&cli.PathFlag{
				Name:        "log",
				Value:       "chrootd.log",
				EnvVars:     []string{"CHROOTD_LOGFILE"},
				Usage:       "daemon log path",
				Destination: &u.LogFileName,
			},
			&cli.BoolFlag{
				Name:        "rootless",
				Usage:       "if runs in rootless mode",
				Value:       true,
				Destination: &u.Rootless,
			},
			&cli.StringFlag{
				Name:  "loglevel",
				Value: "info",
				Usage: "set log level: debug, info, warn, error",
			},
			&cli.IntFlag{
				Name:  "proclimits",
				Value: 64,
				Usage: "maximum attachable proccess limits",
			},
			&cli.BoolFlag{
				Name:  "daemon",
				Value: false,
				Usage: "start in background",
			},
		},
		Before: utils.NewTomlFlagLoader("config"),
		Action: func(c *cli.Context) error {
			user := c.Context.Value("_data").(*User)

			if c.Bool("daemon") {
				return ErrDaemonStart
			}

			if !filepath.IsAbs(user.RunPath) {
				return errors.New("runpath should be absolute")
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

			switch c.String("regback") {
			case "bolt":
				store, err := libkv.NewStore(store.BOLTDB, c.StringSlice("registry"), &store.Config{Bucket: c.String("registry_bucket")})
				if err != nil {
					return err
				}
				defer store.Close()

				registry = cntr.NewStoreRegistry(store)
			case "consul":
				store, err := libkv.NewStore(store.CONSUL, c.StringSlice("registry"), &store.Config{})
				if err != nil {
					return err
				}
				defer store.Close()

				registry = cntr.NewStoreRegistry(store)
			case "etcd":
				store, err := libkv.NewStore(store.ETCD, c.StringSlice("registry"), &store.Config{})
				if err != nil {
					return err
				}
				defer store.Close()

				registry = cntr.NewStoreRegistry(store)
			case "zk":
				store, err := libkv.NewStore(store.ZK, c.StringSlice("registry"), &store.Config{})
				if err != nil {
					return err
				}
				defer store.Close()

				registry = cntr.NewStoreRegistry(store)
			}

			srv := server.NewServer()

			states, err := libkv.NewStore(store.BOLTDB, []string{filepath.Join(user.RunPath, "states")}, &store.Config{Bucket: "states"})
			if err != nil {
				return err
			}
			defer states.Close()

			cntrServer, err := cntr.NewServer(filepath.Join(user.RunPath, "cntrs"),
				"127.0.0.1:9091",
				states,
				registry)
			if err != nil {
				return err
			}
			defer cntrServer.Close()

			err = cntrServer.Register(srv, user.ServicePath)
			if err != nil {
				return err
			}

			user.Logger.Log().Msgf("rpcx server started at %s", user.Addr)

			go func() {
				h := make(chan os.Signal, 1)
				signal.Notify(h, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
				s := <-h
				switch s {
				case syscall.SIGKILL:
					srv.Close()
				default:
					srv.Shutdown(c.Context)
				}
			}()

			go func() {
				lis := net.ListenConfig{}

				listener, err := lis.Listen(c.Context, "tcp", ":9091")
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

			ua := utils.NewAddrFromString(user.Addr)
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
