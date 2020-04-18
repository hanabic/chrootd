package main

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/api"
	"google.golang.org/grpc/metadata"
)

var ImgUpload = &cli.Command{
	Name:  "upload",
	Usage: "upload a image",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "name",
			Aliases:  []string{"n"},
			Usage:    "the image name",
			Required: true,
		},
		&cli.PathFlag{
			Name:     "meta",
			Aliases:  []string{"m"},
			Usage:    "metajson",
			Required: true,
		},
		&cli.PathFlag{
			Name:     "rootfs",
			Aliases:  []string{"r"},
			Usage:    "rootfs",
			Required: true,
		},
	},
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		content, err := ioutil.ReadFile(c.Path("meta"))
		if err != nil {
			return err
		}

		ctx := metadata.NewOutgoingContext(c.Context, metadata.Pairs(
			"name", c.String("name"),
			"meta-bin", string(content),
		))

		client := api.NewImageClient(user.Conn)

		stream, err := client.Upload(ctx)
		if err != nil {
			return err
		}
		defer stream.CloseSend()

		fd, err := os.Open(c.Path("rootfs"))
		if err != nil {
			return err
		}
		defer fd.Close()

		buf := make([]byte, 1024)
		for {
			n, err := fd.Read(buf)
			if err != nil {
				if err == io.EOF {
					break
				}

				return err
			}

			if err := stream.Send(&api.ImageUploadReq{
				D: buf[:n],
			}); err != nil {
				user.Logger.Err(err).Msgf("terminated by server, bad upload behavior")
				break
			}
		}

		_, err = stream.CloseAndRecv()
		return err
	},
}
