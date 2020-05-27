package main

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var ImgRemove = &cli.Command{
	Name:      "remove",
	Usage:     "remove the specific roofs of the specific metadata, or all rootfs if no id of rootfs is provided",
	Aliases:   []string{"rm", "r"},
	ArgsUsage: "$metaid [$rootfsid]",
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		if c.Args().Len() < 1 {
			return errors.New("must specify at least one argument")
		}

		sli := c.Args().Slice()

		if len(sli) > 1 {
			err := user.Meta.ImageDelete(sli[0], sli[1])
			if err != nil {
				return err
			}
			return nil
		}

		meta, err := user.Meta.Get(sli[0])
		if err != nil {
			return err
		}

		for _, m := range meta.RootfsIds {
			err := user.Meta.ImageDelete(sli[0], m)
			if err != nil {
				return err
			}
		}

		return nil
	},
}
