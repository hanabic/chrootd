package main

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var ImgUnpack = &cli.Command{
	Name:      "unpack",
	Aliases:   []string{"u"},
	Usage:     "unpack a rootfs according to the metadata",
	ArgsUsage: "$metaid",
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		if c.Args().Len() < 1 {
			return errors.New("must specify at least one argument")
		}

		rid, err := user.Meta.ImageUnpack(c.Context, c.Args().First())
		if err != nil {
			return err
		}

		user.Logger.Info().Msgf("rootfs id is %s", rid)

		return nil
	},
}
