package main

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/api"
	"github.com/xhebox/chrootd/utils"
)

var CntrCreate = &cli.Command{
	Name:  "create",
	Usage: "create a container",
	Flags: append(flags,
		&cli.StringFlag{
			Name:  "name",
			Value: "hello",
			Usage: "containere name",
		},
		&cli.StringFlag{
			Name:     "image",
			Usage:    "image name",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "config",
			Value: "./base.toml",
			Usage: "base config",
		}),
	Before: utils.NewTomlFlagLoader("config"),
	Action: func(c *cli.Context) error {
		data := c.Context.Value("_data").(*User)

		client := api.NewContainerClient(data.Conn)

		m, err := MetaFromCli(c)
		if err != nil {
			return errors.Wrapf(err, "fail to make meta")
		}

		cfg, err := m.ToBytes()
		if err != nil {
			return errors.Wrapf(err, "fail to create")
		}

		res, err := client.Create(c.Context, &api.CntrCreateReq{
			Id:     "",
			Config: cfg,
		})
		if err != nil {
			return errors.Wrapf(err, "fail to create")
		}

		out := bytes.NewBufferString("")

		if err := json.Indent(out, res.Config, " ", "\t"); err != nil {
			return errors.Wrapf(err, "fail to marshal indent")
		}

		data.Logger.Info().Msgf("created container[%s]:\n%s", res.Id, out.String())

		return nil
	},
}
