package main

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var Create = &cli.Command{
	Name:      "create",
	Usage:     "create a container based on the specific metadata, rootfs automatically unpacked, a convenient wrapper",
	ArgsUsage: "$metaid",
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		if c.Args().Len() < 1 {
			return errors.New("must specify at least one argument")
		}

		id := c.Args().First()

		meta, err := user.Meta.Get(id)
		if err != nil {
			return err
		}

		rid, err := user.Meta.ImageUnpack(c.Context, id)
		if err != nil {
			return err
		}

		cid, err := user.Cntr.Create(meta, rid)
		if err != nil {
			return err
		}

		user.Logger.Info().Msgf("container id is %s", cid)

		return nil
	},
}
