package pool

import (
	"fmt"

	"github.com/segmentio/ksuid"
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/api"
	. "github.com/xhebox/chrootd/cli/types"
)

func metaFromCli(c *cli.Context) *api.Metainfo {
	return &api.Metainfo{
		Name:   c.String("name"),
		Rootfs: c.String("rootfs"),
	}
}

var Create = &cli.Command{
	Name:  "create",
	Usage: "create a container",
	Flags: []cli.Flag{
		// TODO: more fields
		&cli.StringFlag{
			Name:  "name",
			Value: "hello",
			Usage: "containere name",
		},
		&cli.StringFlag{
			Name:  "rootfs",
			Value: "/",
			Usage: "containere rootfs",
		},
	},
	Action: func(c *cli.Context) error {
		data := c.Context.Value("_data").(*User)

		client := api.NewContainerPoolClient(data.Conn)

		m := metaFromCli(c)

		cfg, err := m.ToBytes()
		if err != nil {
			return fmt.Errorf("fail to create: %s", err)
		}

		res, err := client.Create(c.Context, &api.CreateReq{Config: cfg})
		if err != nil {
			return fmt.Errorf("fail to create: %s", err)
		}
		if len(res.Reason) != 0 {
			return fmt.Errorf("fail to create: %s", res.Reason)
		}

		id, _ := ksuid.FromBytes(res.Id)
		data.Logger.Info().Msgf("created container[%s]", id)

		return nil
	},
}
