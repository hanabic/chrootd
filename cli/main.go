package main

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/registry"
	"github.com/xhebox/chrootd/utils"
)

type User struct {
	Logger zerolog.Logger

	ServicePath         string
	ServiceReadTimeout  time.Duration
	ServiceWriteTimeout time.Duration

	Discovery registry.Discovery
	Registry  registry.Registry
}

func main() {
	u := &User{
		Logger: zerolog.New(os.Stdout),
	}

	app := &cli.App{
		UseShortOptionHandling: true,
		Flags: utils.ConcatMultipleFlags(utils.ZerologFlags,
			registry.RegistryFlags,
			[]cli.Flag{
				&cli.StringFlag{
					Name:        "service_path",
					Usage:       "`service` path used by rpcx",
					Value:       "cntr",
					Destination: &u.ServicePath,
				},
				&cli.StringSliceFlag{
					Name:        "service_addr",
					Usage:       "non-empty value means no service discovery; if more than one address(cluser), a registry is needed for container discovery",
					Value:       cli.NewStringSlice("tcp@:9090"),
					//Destination: &u.ServiceAddrs,
				},
				&cli.DurationFlag{
					Name:        "service_readtimeout",
					Usage:       "server read `timeout`",
					Value:       3 * time.Second,
					Destination: &u.ServiceReadTimeout,
				},
				&cli.DurationFlag{
					Name:        "service_writetimeout",
					Usage:       "server write `timeout`",
					Value:       3 * time.Second,
					Destination: &u.ServiceWriteTimeout,
				},
			}),
		Commands: cli.Commands{
			CntrCreate,
			/*
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
			*/
		},
		Before: func(c *cli.Context) error {
			user := c.Context.Value("_data").(*User)

			var err error
			user.Logger, err = utils.NewLogger(c, user.Logger)
			if err != nil {
				return err
			}

			addrs := c.StringSlice("service_addr")
			if len(addrs) == 1 {
				user.Discovery = registry.NewPeer(addrs[0])
			} else {
				user.Registry, err = registry.NewRegistryFromCli(c)
				if err != nil {
					return err
				}

				if len(addrs) < 1 {
					user.Discovery = registry.NewWrapRegistry("cntrs", user.Registry)
					user.Registry = registry.NewWrapRegistry("service", user.Registry)
				} else {
					user.Discovery, err = registry.NewMultiple(addrs...)
					if err != nil {
						return err
					}
				}
			}

			return nil
		},
		After: func(c *cli.Context) error {
			user := c.Context.Value("_data").(*User)

			if user.Registry != nil {
				user.Registry.Close()
			}

			return nil
		},
	}

	ctx := context.WithValue(context.Background(), "_data", u)

	if err := app.RunContext(ctx, os.Args); err != nil {
		u.Logger.Fatal().Msg(err.Error())
	}
}
