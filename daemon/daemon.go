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

	"github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/smallnest/rpcx/log"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/server"
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/client"
	cloc "github.com/xhebox/chrootd/cntr/local"
	cpro "github.com/xhebox/chrootd/cntr/proxy"
	mloc "github.com/xhebox/chrootd/meta/local"
	mpro "github.com/xhebox/chrootd/meta/proxy"
	"github.com/xhebox/chrootd/store"
	"github.com/xhebox/chrootd/utils"
)

var (
	ErrDaemonStart = errors.New("daemon start")
)

type User struct {
	Logger zerolog.Logger

	OAuthURL string

	ConsulAddr          string
	ServiceAddr         string
	ServiceHTTP         string
	ServiceReadTimeout  time.Duration
	ServiceWriteTimeout time.Duration
	ServiceRootless     bool
	AttachAddr          string

	ConfPath  string
	RunPath   string
	ImagePath string
}

func main() {
	u := &User{
		Logger: zerolog.New(os.Stdout),
	}

	app := &cli.App{
		Usage:                  "chrootd daemon program",
		EnableBashCompletion:   true,
		UseShortOptionHandling: true,
		Flags: utils.ConcatMultipleFlags(
			[]cli.Flag{
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
					Usage:       "daemon `RUNPATH` for states,unpacked images(should be large enough)",
					Destination: &u.RunPath,
				},
				&cli.StringFlag{
					Name:        "image",
					Value:       "/var/lib/chrootd/images",
					EnvVars:     []string{"CHROOTD_IMGPATH"},
					Usage:       "where oci-layout images stored",
					Destination: &u.ImagePath,
				},
			},
			utils.ZerologFlags,
			[]cli.Flag{
				&cli.StringFlag{
					Name:        "oauth_validate",
					Usage:       "a validation url to verify the token passed by clients, will enable permission control",
					Destination: &u.OAuthURL,
				},
				&cli.StringFlag{
					Name:        "consul_addr",
					Usage:       "`address` for consul agent, will register services online",
					Destination: &u.ConsulAddr,
				},
				&cli.StringFlag{
					Name:        "service_addr",
					Usage:       "`address` to listen container service",
					Value:       ":9090",
					EnvVars:     []string{"CHROOTD_ADDR"},
					Destination: &u.ServiceAddr,
				},
				&cli.StringFlag{
					Name:        "service_http",
					Value:       ":9091",
					Usage:       "will enable an http gateway",
					EnvVars:     []string{"CHROOTD_HTTP"},
					Destination: &u.ServiceHTTP,
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
				&cli.StringFlag{
					Name:        "attach_addr",
					Usage:       "`address` for process attach",
					Value:       ":9092",
					Destination: &u.AttachAddr,
				},
			}),
		Commands: cli.Commands{
			cloc.InitFlag,
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

			states, err := store.NewBolt(filepath.Join(user.RunPath, "states"), "chrootd")
			if err != nil {
				return err
			}
			defer states.Close()

			mmgr, err := mloc.NewMetaManager(user.RunPath, user.ImagePath, states, func(m *mloc.MetaManager) error {
				m.Rootless = user.ServiceRootless
				return nil
			})
			if err != nil {
				return err
			}
			defer mmgr.Close()

			cmgr, err := cloc.NewCntrManager(user.RunPath, user.ImagePath, states, func(m *cloc.CntrManager) error {
				m.Rootless = user.ServiceRootless
				return nil
			})
			if err != nil {
				return err
			}
			defer cmgr.Close()

			errch := make(chan error, 1)

			rpcAddr := utils.NewAddrString("tcp", user.ServiceAddr)
			httpAddr := utils.NewAddrString("tcp", user.ServiceHTTP)
			attachAddr := utils.NewAddrString("tcp", user.AttachAddr)

			srv := server.NewServer(
				server.WithReadTimeout(user.ServiceReadTimeout),
				server.WithWriteTimeout(user.ServiceWriteTimeout),
				func(srv *server.Server) {
					srv.DisableHTTPGateway = true
					srv.DisableJSONRPC = true
				},
			)

			if c.IsSet("oauth_validate") {
				srv.AuthFunc = func(ctx context.Context, req *protocol.Message, token string) error {
					client := http.Client{}

					vreq, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s?access_token=%s&method=%s", user.OAuthURL, token, req.ServiceMethod), nil)
					if err != nil {
						return err
					}

					resp, err := client.Do(vreq)
					if err != nil {
						return err
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						return errors.New("access denied")
					}

					return nil
				}
			}

			var con *api.Client
			if len(user.ConsulAddr) != 0 {
				con, err = api.NewClient(&api.Config{Address: user.ConsulAddr})
				if err != nil {
					return err
				}
			}

			msvc, err := mpro.NewMetaService(mmgr, con, "meta", rpcAddr)
			if err != nil {
				return err
			}

			err = srv.RegisterName("meta", msvc, "")
			if err != nil {
				return err
			}

			csvc, err := cpro.NewCntrService(cmgr, con, "cntr", rpcAddr, attachAddr)
			if err != nil {
				return err
			}

			err = srv.RegisterName("cntr", csvc, "")
			if err != nil {
				return err
			}

			lnRPC, err := net.Listen(rpcAddr.Network(), rpcAddr.String())
			if err != nil {
				return err
			}
			defer lnRPC.Close()

			go func() {
				errch <- srv.ServeListener(rpcAddr.Network(), lnRPC)
			}()

			user.Logger.Info().Msgf("service server started at %s", user.ServiceAddr)

			hsrv := &http.Server{}

			lnHTTP, err := net.Listen(httpAddr.Network(), httpAddr.String())
			if err != nil {
				return err
			}
			defer lnHTTP.Close()

			gateway, err := client.NewGateway(rpcAddr.Network(), rpcAddr.String())
			if err != nil {
				return err
			}

			hsrv.Handler = gateway

			go func() {
				errch <- hsrv.Serve(lnHTTP)
			}()

			user.Logger.Info().Msgf("service http gateway started at %s", user.ServiceHTTP)

			lnAttach, err := net.Listen(attachAddr.Network(), attachAddr.String())
			if err != nil {
				return err
			}
			defer lnAttach.Close()

			go func() {
				errch <- csvc.ServeListener(lnAttach)
			}()

			user.Logger.Info().Msgf("task attach server started at %s", user.AttachAddr)

			h := make(chan os.Signal, 1)

			signal.Notify(h, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

			select {
			case s := <-h:
				if s == syscall.SIGKILL {
					hsrv.Close()
					srv.Close()
					csvc.Shutdown()
					return errors.New("aborted")
				}
			case <-errch:
			}

			csvc.Shutdown()
			hsrv.Shutdown(context.Background())
			srv.Shutdown(context.Background())
			return nil
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
