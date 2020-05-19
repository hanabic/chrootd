package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var TaskList = &cli.Command{
	Name:      "list",
	Usage:     "list all tasks of a container",
	Aliases:   []string{"ls", "l"},
	ArgsUsage: "$cntrid",
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		if c.Args().Len() < 1 {
			return errors.New("must specify at least one argument")
		}

		cntr, err := user.Cntr.Get(c.Args().First())
		if err != nil {
			return err
		}

		err = cntr.List(func(tid string) error {
			fmt.Println(tid)
			return nil
		})
		if err != nil {
			return err
		}

		return nil
	},
}
