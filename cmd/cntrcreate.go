package main

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	ctyp "github.com/xhebox/chrootd/cntr"
)

var CntrCreate = &cli.Command{
	Name:      "create",
	Usage:     "create container based on metadata and rootfs",
	ArgsUsage: "$metaid $rootfsid",
	Aliases:   []string{"c"},
	Flags: []cli.Flag{
		&cli.StringSliceFlag{
			Name:    "tag",
			Aliases: []string{"t"},
			Usage:   "tag containers by string",
		},
	},
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		if c.Args().Len() < 2 {
			return errors.New("must specify at least two arguments")
		}

		args := c.Args().Slice()

		meta, err := user.Meta.Get(args[0])
		if err != nil {
			return err
		}

		cid, err := user.Cntr.Create(&ctyp.Cntrinfo{
			Meta:   meta,
			Rootfs: args[1],
			Tags:   c.StringSlice("tag"),
		})
		if err != nil {
			return err
		}

		user.Logger.Info().Msgf("container id is %s", cid)

		return nil
	},
}
