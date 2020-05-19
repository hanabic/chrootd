package main

import (
	"github.com/urfave/cli/v2"
)

var MetaDelete = &cli.Command{
	Name:      "delete",
	Usage:     "delete metadata",
	Aliases:   []string{"d"},
	ArgsUsage: "[$metaid1, ..., $metaidX]",
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		for _, m := range c.Args().Slice() {
			err := user.Meta.Delete(m)
			if err != nil {
				return err
			}
			user.Logger.Info().Msgf("deleted %s", m)
		}

		return nil
	},
}
