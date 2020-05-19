package main

import (
	"github.com/urfave/cli/v2"
	ctyp "github.com/xhebox/chrootd/cntr"
	"github.com/xhebox/chrootd/utils"
)

var Start = &cli.Command{
	Name:  "start",
	Usage: "start a container based on the specific metadata, rootfs automatically unpacked, a convenient wrapper",
	Flags: utils.ConcatMultipleFlags(
		[]cli.Flag{
			&cli.StringSliceFlag{
				Name:    "tag",
				Aliases: []string{"t"},
				Usage:   "tag containers by string, later can be used to stop/delete by this",
			},
			&cli.UintFlag{
				Name:    "number",
				Aliases: []string{"n", "num"},
				Value:   1,
				Usage:   "the number of container instance to start",
			},
			&cli.BoolFlag{
				Name:    "rdroot",
				Value:   false,
				Aliases: []string{"r"},
				Usage:   "if the rootfs is readonly, useful when you want to create multiple instances sharing one rootfs",
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
	ArgsUsage: "$metaid [args]",
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		id := c.String("id")

		meta, err := user.Meta.Get(id)
		if err != nil {
			return err
		}

		var task *ctyp.Taskinfo
		if c.Args().Len() > 0 {
			task, err = TaskFromCli(c)
			if err != nil {
				return err
			}
		}

		var rid string

		for v, e := uint(0), c.Uint("number"); v < e; v++ {
			if !c.Bool("rdroot") || rid == "" {
				rid, err = user.Meta.ImageUnpack(c.Context, id)
				if err != nil {
					return err
				}
			}

			cid, err := user.Cntr.Create(&ctyp.Cntrinfo{
				Meta:   meta,
				Rootfs: rid,
				Tags:   c.StringSlice("tag"),
			})
			if err != nil {
				return err
			}

			user.Logger.Info().Msgf("started container %s", cid)

			if task != nil {
				cntr, err := user.Cntr.Get(cid)
				if err != nil {
					return err
				}

				tid, err := cntr.Start(task)
				if err != nil {
					return err
				}

				user.Logger.Info().Msgf("started task %s", tid)
			}
		}

		return nil
	},
}
