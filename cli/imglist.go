package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/api"
)

var ImgList = &cli.Command{
	Name:  "list",
	Usage: "list images",
	Flags: []cli.Flag{
		&cli.PathFlag{
			Name:    "output",
			Usage:   "write metajson to a specific dir",
			Aliases: []string{"o"},
		},
	},
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		client := api.NewImageClient(user.Conn)

		stream, err := client.List(c.Context, &api.ImageListReq{})
		if err != nil {
			return err
		}
		defer stream.CloseSend()

		if c.IsSet("output") {
			os.MkdirAll(c.Path("output"), 0755)
		}
		out := bytes.NewBufferString("")
		for {
			img, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				return errors.Wrapf(err, "fail to list images")
			}

			out.Reset()
			if err := json.Indent(out, img.MetadataJson, " ", "\t"); err != nil {
				return errors.Wrapf(err, "fail to marshal indent")
			}

			if c.IsSet("output") {
				user.Logger.Info().Msgf("%s", img.Name)

				if err := ioutil.WriteFile(filepath.Join(c.String("output"), img.Name + ".json"), out.Bytes(), 0644); err != nil {
					return err
				}
			} else {
				user.Logger.Info().Msgf("[%s]:\n%s", img.Name, out.String())
			}
		}

		return nil
	},
}
