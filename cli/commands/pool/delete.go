package pool

import (
	"fmt"

	"github.com/segmentio/ksuid"
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/api"
	. "github.com/xhebox/chrootd/cli/types"
)

var Delete = &cli.Command{
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

		client := api.NewContainerPoolClient(data.Conn)

		id, err := ksuid.Parse(c.String("id"))
		if err != nil {
			return err
		}

		res, err := client.Delete(c.Context, &api.DeleteReq{Id: id.Bytes()})
		if err != nil {
			return fmt.Errorf("fail to delete container[%s]: %s", id, err)
		}
		if len(res.Reason) != 0 {
			return fmt.Errorf("fail to delete container[%s]: %s", id, res.Reason)
		}

		data.Logger.Info().Msgf("deleted container[%s]", id)

		return nil
	},
}
