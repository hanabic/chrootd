package task

import (
	"fmt"
	"io"

	"github.com/segmentio/ksuid"
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/api"
	. "github.com/xhebox/chrootd/cli/types"
)

var ListProc = &cli.Command{
	Name:  "proc",
	Usage: "list processes of a task",
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

		stream, err := client.ListProc(c.Context, &api.ListProcReq{Id: id.Bytes()})
		if err != nil {
			return err
		}

		for {
			proc, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("%v.ListFeatures(_) = _, %v", client, err)
			}

			// TODO: handwritten proc info output, maybe add to utils
			data.Logger.Info().Msgf("task[%s]: %v", id, proc.Info)
		}

		return nil
	},
}
