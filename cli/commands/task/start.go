package task

import (
	"fmt"

	"github.com/segmentio/ksuid"
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/api"
	. "github.com/xhebox/chrootd/cli/types"
)

var Start = &cli.Command{
	Name:  "start",
	Usage: "start a task by container metadata",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "id",
			Usage:    "containere id",
			Required: true,
		},
	},
	Action: func(c *cli.Context) error {
		data := c.Context.Value("_data").(*User)

		client := api.NewTaskClient(data.Conn)

		id, err := ksuid.Parse(c.String("id"))
		if err != nil {
			return err
		}

		res, err := client.Start(c.Context, &api.StartReq{CntrId: id.Bytes()})
		if err != nil {
			return fmt.Errorf("fail to start container[%s]: %s", id, err)
		}
		if len(res.Reason) != 0 {
			return fmt.Errorf("fail to start container[%s]: %s", id, res.Reason)
		}

		tid, _ := ksuid.FromBytes(res.Id)
		data.Logger.Info().Msgf("started container[%s]: task - %s", id, tid)

		return nil
	},
}
