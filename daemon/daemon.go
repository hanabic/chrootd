package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/smallnest/rpcx/log"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/serverplugin"
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/cntr"
	"github.com/xhebox/chrootd/registry"
	"github.com/xhebox/chrootd/utils"
	"github.com/xhebox/libkv-bolt"
)

func init() {
	boltdb.Register()
}

var (
	ErrDaemonStart = errors.New("daemon start")
)

type User struct {
	Logger zerolog.Logger

	OAuthURL string

	ServiceAddr         string
	ServicePath         string
	ServiceReadTimeout  time.Duration
	ServiceWriteTimeout time.Duration
	ServiceRootless     bool
	ServiceSecure       bool

	AttachAddr   string
	AttachLimits int

	ImageAddr string

	ConfPath string
	RunPath  string
}

func main() {
	u := &User{
		Logger: zerolog.New(os.Stdout),
	}

	app := &cli.App{
		Usage:                  "chrootd daemon program",
		EnableBashCompletion:   true,
		UseShortOptionHandling: true,
		Flags: utils.ConcatMultipleFlags([]cli.Flag{
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
				Name:        "run",
				Value:       "/var/lib/chrootd",
				EnvVars:     []string{"CHROOTD_RUNPATH"},
				Usage:       "daemon `RUNPATH` for states/persistence",
				Destination: &u.RunPath,
			},
		},
			utils.ZerologFlags,
			registry.RegistryFlags,
			[]cli.Flag{
				&cli.StringFlag{
					Name:        "oauth_validate",
					Usage:       "a validation url to verify the token passed by clients, will enable permission control",
					Destination: &u.OAuthURL,
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
				&cli.StringFlag{
					Name:        "image_addr",
					Usage:       "`address` for image uploading",
					Value:       "tcp@:9092",
					Destination: &u.ImageAddr,
				},
			}),
		Before: utils.NewTomlFlagLoader("config"),
		Action: func(c *cli.Context) error {
			user := c.Context.Value("_data").(*User)

			if c.Bool("daemon") {
				return ErrDaemonStart
			}

			if !filepath.IsAbs(user.RunPath) {
				return errors.New("run should be absolute")
			}

			err := os.MkdirAll(user.RunPath, 0755)
			if err != nil {
				return err
			}

			user.Logger, err = utils.NewLogger(c, user.Logger)
			if err != nil {
				return err
			}

			user.Logger.Info().Msgf("daemon started, logleve - %s", user.Logger.GetLevel())

			log.SetLogger(utils.NewRpcxLogger(user.Logger))

			reg, err := registry.NewRegistryFromCli(c)
			if err == nil {
				defer reg.Close()
			} else {
				user.Logger.Err(err).Msgf("fail to start registry")
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
				registry.NewWrapRegistry("service", reg),
				func(s *cntr.Server) error {
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

			imageServer := serverplugin.NewFileTransfer(user.ImageAddr, cntrServer.SaveImage, nil, 1000)

			srv := server.NewServer(
				server.WithReadTimeout(user.ServiceReadTimeout),
				server.WithWriteTimeout(user.ServiceWriteTimeout),
			)
			serverplugin.RegisterFileTransfer(srv, imageServer)

			if c.IsSet("oauth_validate") {
				srv.AuthFunc = func(ctx context.Context, req *protocol.Message, token string) error {
					resp, err := http.Get(fmt.Sprintf("%s?token=%s&method=%s", user.OAuthURL, token, req.ServiceMethod))
					if err != nil {
						return err
					}
					defer resp.Body.Close()
					if resp.StatusCode != http.StatusOK {
						return errors.New("permission denied")
					}
					return nil
				}
			}

			err = cntrServer.Register(srv, user.ServicePath)
			if err != nil {
				return err
			}

			user.Logger.Info().Msgf("container service server started at %s", user.ServiceAddr)
			user.Logger.Info().Msgf("process attach server started at %s", user.AttachAddr)
			user.Logger.Info().Msgf("image uploading server started at %s", user.ImageAddr)

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
			err = srv.Serve(ua.Network(), ua.Addr())
			if err == server.ErrServerClosed {
				return nil
			}
			return err
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
