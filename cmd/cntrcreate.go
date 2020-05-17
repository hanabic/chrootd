package main

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var CntrCreate = &cli.Command{
	Name:      "create",
	Usage:     "create container based on metadata and rootfs",
	ArgsUsage: "$metaid $rootfsid",
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		if c.Args().Len() < 2 {
			return errors.New("must specify at least two arguments")
		}

		args := c.Args().Slice()

		meta, err := user.Meta.Get(args[0])
		if err != nil {
			return err
		}

		cid, err := user.Cntr.Create(meta, args[1])
		if err != nil {
			return err
		}

		user.Logger.Info().Msgf("container id is %s", cid)

		return nil
	},
}
