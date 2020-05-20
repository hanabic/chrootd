package main

import (
	"github.com/urfave/cli/v2"
	ctyp "github.com/xhebox/chrootd/cntr"
	"github.com/xhebox/chrootd/utils"
)

var Exec = &cli.Command{
	Name:      "exec",
	Usage:     "create a container based on the specific metadata, execute a program, attach to it, everything will be released after the termination",
	ArgsUsage: "$metaid",
	Flags: utils.ConcatMultipleFlags(
		[]cli.Flag{
			&cli.BoolFlag{
				Name:    "attach",
				Value:   false,
				Aliases: []string{"a"},
				Usage:   "attach to the process",
			},
			&cli.StringFlag{
				Name:     "id",
				Aliases:  []string{"i"},
				Required: true,
				Usage:    "id of metadata",
			},
		},
		taskFlags,
	),
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		id := c.String("id")

		rid, err := user.Meta.ImageUnpack(c.Context, id)
		if err != nil {
			return err
		}
		defer user.Meta.ImageDelete(id, rid)

		meta, err := user.Meta.Get(id)
		if err != nil {
			return err
		}

		cid, err := user.Cntr.Create(&ctyp.Cntrinfo{
			Meta:   meta,
			Rootfs: rid,
		})
		if err != nil {
			return err
		}

		cntr, err := user.Cntr.Get(cid)
		if err != nil {
			return err
		}
		defer cntr.StopAll(true)

		task, err := TaskFromCli(c)
		if err != nil {
			return err
		}

		tid, err := cntr.Start(task)
		if err != nil {
			return err
		}

		rw, err := cntr.Attach(tid)
		if err != nil {
			return err
		}
		defer rw.Close()

		err = attach(rw, c.Context)
		if err != nil {
			return err
		}

		err = cntr.Wait()
		if err != nil {
			return err
		}

		err = cntr.StopAll(true)
		if err != nil {
			return err
		}

		err = user.Meta.ImageDelete(id, rid)
		if err != nil {
			return err
		}

		return nil
	},
}
