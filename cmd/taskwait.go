package main

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var TaskWait = &cli.Command{
	Name:      "wait",
	Usage:     "wait all tasks of containers done",
	ArgsUsage: "$cntrid",
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		if c.Args().Len() < 1 {
			return errors.New("must specify at least one argument")
		}

		args := c.Args().Slice()

		cntr, err := user.Cntr.Get(args[0])
		if err != nil {
			return err
		}

		err = cntr.Wait()
		if err != nil {
			return err
		}

		return nil
	},
}
