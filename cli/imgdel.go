package main

import (
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/api"
)

var ImgDel = &cli.Command{
	Name:  "delete",
	Usage: "delete a image",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "name",
			Aliases:  []string{"n"},
			Usage:    "the image name",
			Required: true,
		},
	},
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		client := api.NewImageClient(user.Conn)

		_, err := client.Delete(c.Context, &api.ImageDeleteReq{
			Name: c.String("name"),
		})
		if err != nil {
			return err
		}

		return nil
	},
}
