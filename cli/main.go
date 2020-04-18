package main

import (
	"context"
	"net"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
)

type User struct {
	Logger zerolog.Logger
	Conn   *grpc.ClientConn
}

func emptyFormatter(interface{}) string {
	return ""
}

func main() {
	user := &User{
		Logger: zerolog.New(os.Stdout),
	}

	app := &cli.App{
		UseShortOptionHandling: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "addr",
				Usage:   "server connection addr",
				Value:   "127.0.0.1:9090",
				EnvVars: []string{"CHROOTD_CONNADDR"},
			},
			&cli.StringFlag{
				Name:    "network",
				Usage:   "server connection network type",
				Value:   "tcp",
				EnvVars: []string{"CHROOTD_CONNTYPE"},
			},
			&cli.DurationFlag{
				Name:    "timeout",
				Usage:   "server connection timeout",
				Value:   30 * time.Second,
				EnvVars: []string{"CHROOTD_CONNTIMEOUT"},
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Value:   false,
				Aliases: []string{"v"},
			},
			&cli.BoolFlag{
				Name:    "quite",
				Value:   false,
				Aliases: []string{"q"},
			},
			&cli.BoolFlag{
				Name:    "structLogger",
				Value:   false,
				Aliases: []string{"s"},
				Usage:   "output structed log",
			},
			&cli.BoolFlag{
				Name:  "nocolor",
				Value: false,
				Usage: "no color output",
			},
		},
		Commands: cli.Commands{
			/*
				Start,
				Stop,
				ListTask,
				ListProc,
				Exec,
			*/
			CntrCreate,
			CntrUpdate,
			CntrList,
			CntrDelete,
			&cli.Command{
				Name:  "image",
				Usage: "image related",
				Subcommands: []*cli.Command{
					ImgList,
					ImgDel,
					ImgUpload,
				},
			},
		},
		Before: func(c *cli.Context) error {
			user := c.Context.Value("_data").(*User)

			if c.Bool("structLogger") {
				user.Logger = user.Logger.With().Timestamp().Logger()
			} else {
				user.Logger = user.Logger.Output(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
					w.NoColor = c.Bool("nocolor")
					w.FormatTimestamp = emptyFormatter
					w.FormatLevel = emptyFormatter
				}))
			}

			if c.Bool("verbose") {
				user.Logger = user.Logger.Level(zerolog.DebugLevel)
			} else if c.Bool("quite") {
				user.Logger = user.Logger.Level(zerolog.WarnLevel)
			} else {
				user.Logger = user.Logger.Level(zerolog.InfoLevel)
			}

			addr := c.String("addr")
			network := c.String("network")
			timeout := c.Duration("timeout")

			user.Logger.Debug().Msgf("grpc connect to [%s]%s -- %s", network, addr, timeout)

			var err error
			user.Conn, err = grpc.Dial("new", grpc.WithInsecure(), grpc.WithTimeout(timeout), grpc.WithContextDialer(func(ctx context.Context, target string) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, network, addr)
			}))
			if err != nil {
				return err
			}

			return nil
		},
		After: func(c *cli.Context) error {
			user := c.Context.Value("_data").(*User)

			user.Logger.Debug().Msgf("grpc disconnect")

			user.Conn.Close()

			return nil
		},
	}

	ctx := context.WithValue(context.Background(), "_data", user)

	if err := app.RunContext(ctx, os.Args); err != nil {
		user.Logger.Fatal().Msg(err.Error())
	}
}
