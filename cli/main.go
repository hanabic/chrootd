package main

import (
	"context"
	"net"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	. "github.com/xhebox/chrootd/cli/commands/pool"
	. "github.com/xhebox/chrootd/cli/commands/task"
	. "github.com/xhebox/chrootd/cli/types"
	"google.golang.org/grpc"
)

func main() {
	user := &User{
		Logger: zerolog.New(os.Stdout).With().Timestamp().Logger(),
	}

	// TODO: use pkg/errors for wrapping error, instead of fmt.Errorf
	app := &cli.App{
		// TODO: add version flag
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
		Commands: []*cli.Command{
			List,
			Create,
			Update,
			Delete,
			Start,
			Stop,
			ListTask,
			ListProc,
			Exec,
		},
		Before: func(c *cli.Context) error {
			user := c.Context.Value("_data").(*User)

			if !c.Bool("structLogger") {
				user.Logger = user.Logger.Output(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
					w.TimeFormat = time.RFC3339
					w.NoColor = c.Bool("nocolor")
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
			data := c.Context.Value("_data").(*User)

			data.Logger.Debug().Msgf("grpc disconnect")

			data.Conn.Close()

			return nil
		},
	}

	ctx := context.WithValue(context.Background(), "_data", user)

	if err := app.RunContext(ctx, os.Args); err != nil {
		user.Logger.Fatal().Msg(err.Error())
	}
}
