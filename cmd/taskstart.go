package main

import (
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/utils"
)

var TaskStart = &cli.Command{
	Name:    "run",
	Usage:   "run a process in a specific container",
	Aliases: []string{"r"},
	Flags: utils.ConcatMultipleFlags(
		[]cli.Flag{
			&cli.StringFlag{
				Name:     "id",
				Usage:    "container id",
				Required: true,
			},
		},
		taskFlags,
		capFlags,
	),
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		tinfo, err := TaskFromCli(c)
		if err != nil {
			return err
		}

		cntr, err := user.Cntr.Get(c.String("id"))
		if err != nil {
			return err
		}

		tid, err := cntr.Start(tinfo)
		if err != nil {
			return err
		}

		user.Logger.Info().Msgf("task id is %s", tid)

		return nil
	},
}
