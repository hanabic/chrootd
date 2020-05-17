package main

import (
	"context"
	"os"

	"github.com/hashicorp/consul/api"
	"github.com/rs/zerolog"
	"github.com/smallnest/rpcx/share"
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/client"
	"github.com/xhebox/chrootd/utils"
	"golang.org/x/oauth2/clientcredentials"

	ctyp "github.com/xhebox/chrootd/cntr"
	cpro "github.com/xhebox/chrootd/cntr/proxy"
	mtyp "github.com/xhebox/chrootd/meta"
	mpro "github.com/xhebox/chrootd/meta/proxy"
)

type User struct {
	Logger zerolog.Logger

	ConsulAddr string
	ServerAddr string
	AttachAddr string

	Consul *api.Client
	Client client.Client

	Meta mtyp.Manager
	Cntr ctyp.Manager
}

func main() {
	u := &User{
		Logger: zerolog.New(os.Stdout),
	}

	app := &cli.App{
		Usage:                  "chrootd command line tool",
		UseShortOptionHandling: true,
		Flags: utils.ConcatMultipleFlags(utils.ZerologFlags,
			[]cli.Flag{
				&cli.StringFlag{
					Name:  "oauth_id",
					Usage: "application id",
				},
				&cli.StringFlag{
					Name:  "oauth_secret",
					Usage: "application secret",
				},
				&cli.StringFlag{
					Name:  "oauth_token",
					Usage: "enable permission control if this option is given",
				},
				&cli.StringFlag{
					Name:        "server_addr",
					Usage:       "chrootd server addr",
					Value:       ":9090",
					Destination: &u.ServerAddr,
				},
				&cli.StringFlag{
					Name:        "attach_addr",
					Usage:       "chrootd task attach addr",
					Value:       ":9092",
					Destination: &u.AttachAddr,
				},
				&cli.StringFlag{
					Name:        "consul_addr",
					Usage:       "non-empty value will enable consul",
					Destination: &u.ConsulAddr,
				},
			}),
		Commands: cli.Commands{
			&cli.Command{
				Name:  "meta",
				Usage: "manage metadatas",
				Subcommands: cli.Commands{
					MetaCreate,
					MetaUpdate,
					MetaGet,
					MetaQuery,
					MetaDelete,
					MetaEximg,
					MetaLsimg,
					MetaRmimg,
				},
			},
			&cli.Command{
				Name:  "cntr",
				Usage: "manage containers",
				Subcommands: cli.Commands{
					CntrCreate,
					CntrDelete,
					CntrGet,
					CntrQuery,
				},
			},
			&cli.Command{
				Name:  "task",
				Usage: "manage tasks",
				Subcommands: cli.Commands{
					TaskStart,
					TaskList,
					TaskWait,
					TaskStop,
					TaskAttach,
				},
			},
			MetaIMGList,
			Create,
		},
		Before: func(c *cli.Context) error {
			user := c.Context.Value("_data").(*User)

			var err error
			user.Logger, err = utils.NewLogger(c, user.Logger)
			if err != nil {
				return err
			}

			if len(user.ConsulAddr) != 0 {
				user.Consul, err = api.NewClient(&api.Config{Address: user.ConsulAddr})
				if err != nil {
					return err
				}

				user.Meta, err = mpro.NewMetaProxy("meta", user.Consul)
				if err != nil {
					return err
				}

				user.Cntr, err = cpro.NewCntrProxy("cntr", user.Consul, nil)
				if err != nil {
					return err
				}
			} else {
				user.Client, err = client.NewClient("tcp", user.ServerAddr)
				if err != nil {
					return err
				}

				user.Meta, err = mpro.NewMetaProxy("meta", user.Client)
				if err != nil {
					return err
				}

				attach := utils.NewAddrString("tcp", user.AttachAddr)

				user.Cntr, err = cpro.NewCntrProxy("cntr", user.Client, attach)
				if err != nil {
					return err
				}
			}

			if c.IsSet("oauth_token") {
				cfg := &clientcredentials.Config{
					ClientID:     c.String("oauth_id"),
					ClientSecret: c.String("oauth_secret"),
					TokenURL:     c.String("oauth_token"),
					Scopes:       []string{"chrootd"},
				}

				tok, err := cfg.Token(c.Context)
				if err != nil {
					return err
				}

				c.Context = context.WithValue(c.Context, share.ReqMetaDataKey, map[string]string{
					share.AuthKey: tok.AccessToken,
				})
			}

			return nil
		},
		After: func(c *cli.Context) error {
			user := c.Context.Value("_data").(*User)

			if user.Meta != nil {
				user.Meta.Close()
			}

			if user.Cntr != nil {
				user.Cntr.Close()
			}

			if user.Client != nil {
				user.Client.Close()
			}

			return nil
		},
	}

	ctx := context.WithValue(context.Background(), "_data", u)

	if err := app.RunContext(ctx, os.Args); err != nil {
		u.Logger.Fatal().Msg(err.Error())
	}
}
