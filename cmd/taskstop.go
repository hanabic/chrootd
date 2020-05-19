package main

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var TaskStop = &cli.Command{
	Name:      "stop",
	Usage:     "stop a task, no taskid meaning stop all",
	Aliases:   []string{"s"},
	ArgsUsage: "$cntrid [$taskid]",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "kill",
			Value:   false,
			Aliases: []string{"k"},
			Usage:   "kill the task or not",
		},
	},
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

		if len(args) > 1 {
			err = cntr.Stop(args[1], c.Bool("kill"))
			if err != nil {
				return err
			}
		} else {
			err = cntr.StopAll(c.Bool("kill"))
			if err != nil {
				return err
			}
		}

		return nil
	},
}
