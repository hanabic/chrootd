package main

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/cntr"
	"github.com/xhebox/chrootd/utils"
)

var CntrCreate = &cli.Command{
	Name:  "create",
	Usage: "create a container",
	Flags: utils.ConcatMultipleFlags(flags,
		[]cli.Flag{
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
		}),
	Before: utils.NewTomlFlagLoader("config"),
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		m, err := MetaFromCli(c)
		if err != nil {
			return errors.Wrapf(err, "fail to make meta")
		}

		res := &cntr.CreateRes{}
		err = user.Client.Create(c.Context, &cntr.CreateReq{
			Meta: m,
		}, res)
		if err != nil {
			return errors.Wrapf(err, "fail to create")
		}

		user.Logger.Info().Msgf("created container[%s]:\n%s", res.Id, res.Meta)

		return nil
	},
}
