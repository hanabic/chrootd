package pool

import (
	"fmt"
	"github.com/xhebox/chrootd/cli/options"

	"github.com/segmentio/ksuid"
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/api"
	. "github.com/xhebox/chrootd/cli/types"
)

var Update = &cli.Command{
	Name:  "update",
	Usage: "update a container",
	Flags: []cli.Flag{
		// TODO: more fields
		&cli.StringFlag{
			Name:  "name",
			Value: "undefined",
			Usage: "containere name",
		},
		&cli.StringFlag{
			Name:     "id",
			Usage:    "container id",
			Required: true,
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

		id, err := ksuid.Parse(c.String("id"))
		if err != nil {
			return err
		}

		m, err := options.MetaFromCli(c)
		if err != nil {
			return fmt.Errorf("fail to make meta: %s", err)
		}

		cfg, err := m.ToBytes()
		if err != nil {
			return fmt.Errorf("fail to create: %s", err)
		}

		res, err := client.Update(c.Context, &api.UpdateReq{Id: id.Bytes(), Config: cfg})
		if err != nil {
			return fmt.Errorf("fail to update container[%s]: %w", id, err)
		}

		cntr := res.Config
		if cntr == nil {
			data.Logger.Info().Msgf("fail to update container[%s]: %s", id, res.Reason)
		} else {
			config, _ := api.NewMetaFromBytes(cntr)
			data.Logger.Info().Msgf("updated container[%s]: %s", id, config)
		}

		return nil
	},
}
