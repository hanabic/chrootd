package task

import (
	"fmt"

	"github.com/segmentio/ksuid"
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/api"
	. "github.com/xhebox/chrootd/cli/types"
)

var Stop = &cli.Command{
	Name:  "stop",
	Usage: "stop a task",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "id",
			Usage:    "task id",
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

		res, err := client.Stop(c.Context, &api.StopReq{Id: id.Bytes()})
		if err != nil {
			return fmt.Errorf("fail to stop container[%s]: %s", id, err)
		}
		if len(res.Reason) != 0 {
			return fmt.Errorf("fail to stop container[%s]: %s", id, res.Reason)
		}

		data.Logger.Info().Msgf("stopped task[%s]", id)

		return nil
	},
}
