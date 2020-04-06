package pool

import (
	"fmt"
	"github.com/segmentio/ksuid"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	"github.com/xhebox/chrootd/api"
	"github.com/xhebox/chrootd/cli/options"
	. "github.com/xhebox/chrootd/cli/types"
)

var base = append(
	*options.NewCgroupsOptions(),
	*options.NewCapabilitiesOptions()...,
)

var flags = append(base,
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
	&cli.StringFlag{
		Name:  "config",
		Value: "./base.toml",
		Usage: "base config",
	})

var Create = &cli.Command{
	Name:   "create",
	Usage:  "create a container",
	Flags:  flags,
	Before: altsrc.InitInputSourceWithContext(flags, altsrc.NewTomlSourceFromFlagFunc("config")),
	Action: func(c *cli.Context) error {
		fmt.Println(c.String("Capabilities.bounding"))
		data := c.Context.Value("_data").(*User)

		client := api.NewContainerPoolClient(data.Conn)

		m, err := options.MetaFromCli(c)
		if err != nil {
			return fmt.Errorf("fail to make meta: %s", err)
		}

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
