package pool

import (
	"fmt"
	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/segmentio/ksuid"
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/api"
	. "github.com/xhebox/chrootd/cli/types"
	"io"
	"strconv"
)

func formatResource(r configs.Resources) string {

	s := "memory" + ": " + strconv.FormatInt(r.Memory, 10) + "bytes" + "\n"
	s = s + "MemoryReservation" + ": " + strconv.FormatInt(r.MemoryReservation, 10) + "bytes" + "\n"
	s = s + "MemorySwap" + ": " + strconv.FormatInt(r.MemorySwap, 10) + "bytes" + "\n"
	s = s + "KernelMemory" + ": " + strconv.FormatInt(r.KernelMemory, 10) + "bytes" + "\n"
	s = s + "KernelMemoryTCP" + ": " + strconv.FormatInt(r.KernelMemoryTCP, 10) + "bytes" + "\n"
	s = s + "PidsLimit" + ": " + strconv.FormatInt(r.PidsLimit, 10) + "\n"
	s = s + "CpuRtRuntime" + ": " + strconv.FormatInt(r.CpuRtRuntime, 10) + "us" + "\n"
	s = s + "CpuRtPeriod" + ": " + strconv.FormatUint(r.CpuRtPeriod, 10) + "us" + "\n"
	s = s + "CpuPeriod" + ": " + strconv.FormatUint(r.CpuPeriod, 10) + "us" + "\n"
	s = s + "CpuShares" + ": " + strconv.FormatUint(r.CpuShares, 10) + "us" + "\n"
	s = s + "CpuWeight" + ": " + strconv.FormatUint(r.CpuWeight, 10) + "us" + "\n"
	s = s + "CpuShares" + ": " + r.CpusetMems + "us" + "\n"
	s = s + "CpuWeight" + ": " + r.CpusetCpus + "us" + "\n"
	return s
}

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

			data.Logger.Info().Msgf("container[%s]:\nCgroups\n%v", id, formatResource(*config.Config.Cgroups.Resources))

		}

		return nil
	},
}
