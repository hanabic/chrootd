package pool

import (
	"fmt"

	"github.com/segmentio/ksuid"
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/api"
	. "github.com/xhebox/chrootd/cli/types"
)

func fromCli(c *cli.Context) *api.Container {
	return &api.Container{
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

		res, err := client.Create(c.Context, &api.CreateReq{Container: fromCli(c)})
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
