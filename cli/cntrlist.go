package main

import (
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/cntr"
)

var CntrList = &cli.Command{
	Name:  "list",
	Usage: "list containers",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "id",
			Usage: "list a specific container",
			Value: "",
		},
	},
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		res := &cntr.ListRes{}

		err := user.Client.List(c.Context, &cntr.ListReq{
			Filters: []cntr.ListFilter{
				cntr.ListFilter{
					Key: "id",
					Val: c.String("id"),
				},
				// TODO: add more filters
			},
		}, res)
		if err != nil {
			return err
		}

		ids := res.CntrIds

		maxlen := 0
		for i := range ids {
			if k := len(ids[i].Id); k > maxlen {
				maxlen = k
			}
		}

		t1 := "Container Id"
		t2 := "Service Address"

		user.Logger.Log().Msgf("%s%*s\t%s", t1, maxlen-len(t1), "", t2)
		for _, cntr := range res.CntrIds {
			user.Logger.Log().Msgf("%s%*s\t%s", cntr.Id, maxlen-len(cntr.Id), "", cntr.Addr)
		}

		return nil
	},
}
