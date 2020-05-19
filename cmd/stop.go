package main

import (
	"github.com/urfave/cli/v2"
	ctyp "github.com/xhebox/chrootd/cntr"
)

func stopCntr(user *User, c *cli.Context, v string) (string, error) {
	cntr, err := user.Cntr.Get(v)
	if err != nil {
		return "", err
	}

	meta, err := cntr.Meta()
	if err != nil {
		return "", err
	}

	err = cntr.StopAll(c.Bool("kill"))
	if err != nil {
		return "", err
	}

	user.Logger.Info().Msgf("stopped %s", v)

	if c.Bool("delete") {
		err = user.Cntr.Delete(v)
		if err != nil {
			return "", err
		}

		user.Logger.Info().Msgf("deleted %s", v)
	}

	return meta.Rootfs, nil
}

var Stop = &cli.Command{
	Name:  "stop",
	Usage: "stop containers with the specific id or tag",
	Flags: []cli.Flag{
		&cli.StringSliceFlag{
			Name:    "tag",
			Aliases: []string{"t"},
			Usage:   "tag containers by string, later can be used to stop/delete by this",
		},
		&cli.BoolFlag{
			Name:    "delete",
			Aliases: []string{"d"},
			Value:   false,
			Usage:   "also delete containers",
		},
		&cli.BoolFlag{
			Name:    "kill",
			Aliases: []string{"k"},
			Value:   false,
			Usage:   "kill containers",
		},
		&cli.BoolFlag{
			Name:    "rmimg",
			Aliases: []string{"r"},
			Value:   false,
			Usage:   "also remove rootfs",
		},
	},
	ArgsUsage: "[$cntr1id ... $cntrXid]",
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		rmimg := c.Bool("rmimg")

		if c.Args().Len() > 0 {
			for _, v := range c.Args().Slice() {
				rid, err := stopCntr(user, c, v)
				if err != nil {
					return err
				}

				if rmimg {
					err = user.Meta.ImageDelete(v, rid)
					if err != nil {
						return err
					}
				}
			}
		}

		for _, v := range c.StringSlice("tag") {
			res := []string{}

			err := user.Cntr.List(v, func(info *ctyp.Cntrinfo) error {
				res = append(res, info.Id)
				return nil
			})
			if err != nil {
				return err
			}

			rids := []string{}

			for _, v := range res {
				rid, err := stopCntr(user, c, v)
				if err != nil {
					return err
				}
				rids = append(rids, rid)
			}

			if rmimg {
				for k, v := range res {
					err = user.Meta.ImageDelete(v, rids[k])
					if err != nil {
						return err
					}
				}
			}
		}

		return nil
	},
}
