package main

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var MetaRmimg = &cli.Command{
	Name:      "rmimg",
	Usage:     "rm the specific roofs of the specific metadata",
	ArgsUsage: "$rootfsid",
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		if c.Args().Len() < 2 {
			return errors.New("must specify at least two argument")
		}

		sli := c.Args().Slice()

		err := user.Meta.ImageDelete(sli[0], sli[1])
		if err != nil {
			return err
		}

		return nil
	},
}
