package main

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/api"
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
		data := c.Context.Value("_data").(*User)

		client := api.NewContainerClient(data.Conn)

		stream, err := client.List(c.Context, &api.CntrListReq{
			Id:      c.String("id"),
			Filters: []*api.Filter{
				// TODO: add more filters
			},
		})
		if err != nil {
			return err
		}
		defer stream.CloseSend()

		out := bytes.NewBufferString("")

		for {
			cntr, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				return errors.Wrapf(err, "network error")
			}

			out.Reset()
			if err := json.Indent(out, cntr.Config, " ", "\t"); err != nil {
				return errors.Wrapf(err, "fail to marshal indent")
			}

			data.Logger.Info().Msgf("cntr[%s]:\n%s", cntr.Id, out.String())
		}

		return nil
	},
}
