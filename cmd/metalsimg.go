package main

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var MetaLsimg = &cli.Command{
	Name:      "lsimg",
	Usage:     "list rootfs belong to the specific metadata",
	ArgsUsage: "$metaid",
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		if c.Args().Len() < 1 {
			return errors.New("must specify at least one argument")
		}

		err := user.Meta.ImageList(c.Args().First(), func(rid string) error {
			user.Logger.Log().Msg(rid)
			return nil
		})
		if err != nil {
			return err
		}

		return nil
	},
}
