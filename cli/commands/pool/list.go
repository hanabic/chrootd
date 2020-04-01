package pool

import (
	"fmt"
	"io"

	"github.com/segmentio/ksuid"
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/api"
	. "github.com/xhebox/chrootd/cli/types"
)

var List = &cli.Command{
	Name:  "list",
	Usage: "list a container",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "name",
			Value: "*",
			Usage: "container name",
		},
		&cli.StringFlag{
			Name:  "id",
			Value: "*",
			Usage: "container id",
		},
		&cli.StringFlag{
			Name:  "label",
			Value: "*",
			Usage: "container label",
		},
		&cli.StringFlag{
			Name:  "hostname",
			Value: "*",
			Usage: "container hostname",
		},
	},
	Action: func(c *cli.Context) error {
		data := c.Context.Value("_data").(*User)

		client := api.NewContainerPoolClient(data.Conn)

		stream, err := client.List(c.Context, &api.ListReq{
			Filters: []*api.ListReq_Filter{
				&api.ListReq_Filter{
					Key: "name",
					Val: c.String("name"),
				},
			},
		})
		if err != nil {
			return err
		}

		for {
			cntr, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("%v.ListFeatures(_) = _, %v", client, err)
			}

			id, _ := ksuid.FromBytes(cntr.Id)
			config, _ := api.NewMetaFromBytes(cntr.Config)
			data.Logger.Info().Msgf("container[%s]: %+v", id, config)
		}

		return nil
	},
}
