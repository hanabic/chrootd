package main

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var CntrDelete = &cli.Command{
	Name:      "delete",
	Usage:     "delete a container by id",
	ArgsUsage: "$cntrid",
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		if c.Args().Len() < 1 {
			return errors.New("must specify at least one argument")
		}

		err := user.Cntr.Delete(c.Args().First())
		if err != nil {
			return err
		}

		return nil
	},
}
