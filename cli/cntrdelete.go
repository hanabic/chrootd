package main

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var CntrDelete = &cli.Command{
	Name:  "delete",
	Usage: "delete a container",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "id",
			Usage:    "container id",
			Required: true,
		},
	},
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		err := user.Client.Delete(c.Context, c.String("id"))
		if err != nil {
			return errors.Wrapf(err, "fail to delete container")
		}

		user.Logger.Info().Msgf("deleted container[%s]", c.String("id"))

		return nil
	},
}
