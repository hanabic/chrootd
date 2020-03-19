package task

import (
	"fmt"
	"io"

	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/api"
	. "github.com/xhebox/chrootd/cli/types"
)

var ListTask = &cli.Command{
	Name:  "task",
	Usage: "list tasks",
	Flags: []cli.Flag{},
	Action: func(c *cli.Context) error {
		data := c.Context.Value("_data").(*User)

		client := api.NewTaskClient(data.Conn)

		stream, err := client.List(c.Context, &api.ListReq{
			Filters: []*api.ListReq_Filter{
				// TODO: add more filters
			},
		})
		if err != nil {
			return err
		}

		for {
			task, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("%v.ListFeatures(_) = _, %v", client, err)
			}

			// TODO: handwritten task info output, maybe add to utils
			data.Logger.Info().Msgf("task[%s]: %v", task.Id, task)
		}

		return nil
	},
}
