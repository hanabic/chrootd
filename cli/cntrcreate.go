package main

import (
	"github.com/pkg/errors"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/protocol"
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/cntr"
	"github.com/xhebox/chrootd/utils"
)

var CntrCreate = &cli.Command{
	Name:  "create",
	Usage: "create a container",
	Flags: append(flags,
		&cli.StringFlag{
			Name:  "name",
			Value: "hello",
			Usage: "containere name",
		},
		&cli.StringFlag{
			Name:     "image",
			Usage:    "image name",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "config",
			Value: "./base.toml",
			Usage: "base config",
		}),
	Before: utils.NewTomlFlagLoader("config"),
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		n, err := cntr.NewClient(user.ServicePath,
			user.Discovery,
			user.Registry,
			client.Option{
				ReadTimeout:   user.ServiceReadTimeout,
				WriteTimeout:  user.ServiceWriteTimeout,
				SerializeType: protocol.MsgPack,
			})
		if err != nil {
			return err
		}

		m, err := MetaFromCli(c)
		if err != nil {
			return errors.Wrapf(err, "fail to make meta")
		}

		res := &cntr.CreateRes{}
		err = n.Create(c.Context, &cntr.CreateReq{
			Meta: m,
		}, res)
		if err != nil {
			return errors.Wrapf(err, "fail to create")
		}

		user.Logger.Info().Msgf("created container[%s]:\n%s", res.Id, res.Meta)

		return nil
	},
}
