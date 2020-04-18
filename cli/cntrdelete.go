package main

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/api"
)

var CntrDelete = &cli.Command{
	Name:  "delete",
	Usage: "delete a container",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "id",
			Usage:    "container id",
			Required: true,
		},
	},
	Action: func(c *cli.Context) error {
		data := c.Context.Value("_data").(*User)

		client := api.NewContainerClient(data.Conn)

		_, err := client.Delete(c.Context, &api.CntrDeleteReq{
			Id: c.String("id"),
		})
		if err != nil {
			return errors.Wrapf(err, "fail to delete container")
		}

		data.Logger.Info().Msgf("deleted container[%s]", c.String("id"))

		return nil
	},
}
