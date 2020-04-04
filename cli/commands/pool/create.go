package pool

import (
	"fmt"

	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/segmentio/ksuid"
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/api"
	. "github.com/xhebox/chrootd/cli/types"
	"golang.org/x/sys/unix"
)

func metaFromCli(c *cli.Context) *api.Metainfo {
	defaultMountFlags := unix.MS_NOEXEC | unix.MS_NOSUID | unix.MS_NODEV

	return &api.Metainfo{
		Name: c.String("name"),
		Config: configs.Config{
			Rootfs: c.String("rootfs"),
			Capabilities: &configs.Capabilities{
				Bounding: []string{
					"CAP_SETUID",
					"CAP_SETGID",
				},
				Effective: []string{
					"CAP_SETUID",
					"CAP_SETGID",
				},
				Inheritable: []string{
					"CAP_SETUID",
					"CAP_SETGID",
				},
				Permitted: []string{
					"CAP_SETUID",
					"CAP_SETGID",
				},
				Ambient: []string{
					"CAP_SETUID",
					"CAP_SETGID",
				},
			},
			Cgroups: &configs.Cgroup{
				Name:   "Test",
				Parent: "system",
				Resources: &configs.Resources{
					MemorySwappiness: nil,
					AllowAllDevices:  nil,
					AllowedDevices:   configs.DefaultAllowedDevices,
				},
			},
			Namespaces: configs.Namespaces([]configs.Namespace{
				{Type: configs.NEWUTS},
				{Type: configs.NEWIPC},
				{Type: configs.NEWPID},
				{Type: configs.NEWNET},
				{Type: configs.NEWNS},
				{Type: configs.NEWUSER},
			}),
			Devices: configs.DefaultAutoCreatedDevices,
			Mounts: []*configs.Mount{
				{
					Source:      "proc",
					Destination: "/proc",
					Device:      "proc",
					Flags:       defaultMountFlags,
				},
				{
					Source:      "devtmpfs",
					Destination: "/dev",
					Device:      "tmpfs",
					Flags:       unix.MS_NOSUID | unix.MS_STRICTATIME,
					Data:        "mode=755",
				},
				{
					Source:      "sysfs",
					Destination: "/sys",
					Device:      "sysfs",
					Flags:       defaultMountFlags | unix.MS_RDONLY,
				},
			},
		},
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
