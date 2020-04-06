package options

import (
	"errors"
	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	"github.com/xhebox/chrootd/api"
	"golang.org/x/sys/unix"
	"strconv"
	"strings"
	"unicode"
)

//Todo: rlimit options

func MetaFromCli(c *cli.Context) (*api.Metainfo, error) {
	defaultMountFlags := unix.MS_NOEXEC | unix.MS_NOSUID | unix.MS_NODEV

	memory, err := PaserMemory(c.String("Cgroups.memory"))
	if err != nil {
		return nil, err
	}
	memory_swap, err := PaserMemory(c.String("Cgroups.memory-swap"))
	if err != nil {
		return nil, err
	}
	memory_reservation, err := PaserMemory(c.String("Cgroups.memory-reservation"))
	if err != nil {
		return nil, err
	}
	kernal_memory, err := PaserMemory(c.String("Cgroups.kernel-memory"))
	if err != nil {
		return nil, err
	}

	return &api.Metainfo{
		Name: c.String("name"),
		Config: configs.Config{
			Rootfs: c.String("rootfs"),
			Capabilities: &configs.Capabilities{
				Bounding:    c.StringSlice("Capabilities.bounding"),
				Effective:   c.StringSlice("Capabilities.effective"),
				Inheritable: c.StringSlice("Capabilities.inheritable"),
				Permitted:   c.StringSlice("Capabilities.permitted"),
				Ambient:     c.StringSlice("Capabilities.ambient"),
			},
			Cgroups: &configs.Cgroup{
				Name:   "Test",
				Parent: "system",
				Resources: &configs.Resources{
					MemorySwappiness:  nil,
					AllowAllDevices:   nil,
					AllowedDevices:    configs.DefaultAllowedDevices,
					BlkioWeight:       uint16(c.Int("Cgroups.blkio-weight")),
					CpuShares:         uint64(c.Int("Cgroups.cpu-shares")),
					CpuQuota:          int64(c.Int("Cgroups.cpu-quota")),
					CpuPeriod:         uint64(c.Int("Cgroups.cpu-period")),
					CpuRtRuntime:      int64(c.Int("Cgroups.cpu-rt-quota")),
					CpuRtPeriod:       uint64(c.Int("Cgroups.cpu-rt-period")),
					CpusetCpus:        c.String("Cgroups.cpuset-cpus"),
					CpusetMems:        c.String("Cgroups.cpuset-mems"),
					KernelMemory:      int64(kernal_memory),
					Memory:            int64(memory),
					MemorySwap:        int64(memory_swap),
					MemoryReservation: int64(memory_reservation),
					PidsLimit:         int64(c.Int("Cgroups.pids-limit")),
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
	}, nil
}

func PaserMemory(str string) (int, error) {
	unit := "b"
	num, err := strconv.Atoi(str[0 : len(str)-1])

	if !unicode.IsNumber([]rune(str)[len(str)-1]) {
		if err != nil {
			return 0, errors.New("format error")
		}
		unit = strings.ToLower(str[len(str)-1 : len(str)])
	}

	switch unit {
	case "b":
		return num, nil
	case "k":
		return 1024 * num, nil
	case "m":
		return 1024 * 1024 * num, nil
	case "g":
		return 1024 * 1024 * 1024 * num, nil
	default:
		return 0, errors.New("unit should be `b`, `k`, `m` or `g`")
	}
	return num, nil
}

func NewCapabilitiesOptions() *[]cli.Flag {
	return &[]cli.Flag{
		altsrc.NewStringSliceFlag(&cli.StringSliceFlag{
			Name: "Capabilities.bounding",
			Value: cli.NewStringSlice(
				"CAP_SETUID",
				"CAP_SETGID",
			),
			Aliases: []string{"cap_bou"},
		}),
		altsrc.NewStringSliceFlag(&cli.StringSliceFlag{
			Name: "Capabilities.effective",
			Value: cli.NewStringSlice(
				"CAP_SETUID",
				"CAP_SETGID",
			),
			Aliases: []string{"cap_eff"},
		}),
		altsrc.NewStringSliceFlag(&cli.StringSliceFlag{
			Name: "Capabilities.inheritable",
			Value: cli.NewStringSlice(
				"CAP_SETUID",
				"CAP_SETGID",
			),
			Aliases: []string{"cap_inh"},
		}),
		altsrc.NewStringSliceFlag(&cli.StringSliceFlag{
			Name: "Capabilities.permitted",
			Value: cli.NewStringSlice(
				"CAP_SETUID",
				"CAP_SETGID",
			),
			Aliases: []string{"cap_per"},
		}),
		altsrc.NewStringSliceFlag(&cli.StringSliceFlag{
			Name: "Capabilities.ambient",
			Value: cli.NewStringSlice(
				"CAP_SETUID",
				"CAP_SETGID",
			),
			Aliases: []string{"cap_amb"},
		}),
	}
}

func NewCgroupsOptions() *[]cli.Flag {
	return &[]cli.Flag{
		altsrc.NewIntFlag(&cli.IntFlag{
			Name:    "Cgroups.blkio-weight",
			Value:   0,
			Usage:   "Block IO (relative weight), between 10 and 1000, or 0 to disable (default 0)",
			Aliases: []string{"blkio-weight"},
		}),
		altsrc.NewIntFlag(&cli.IntFlag{
			Name:    "Cgroups.cpu-shares",
			Value:   0,
			Usage:   "CPU shares (relative weight vs. other containers",
			Aliases: []string{"cpu-shares"},
		}),
		altsrc.NewIntFlag(&cli.IntFlag{
			Name:    "Cgroups.cpu-quota",
			Value:   0,
			Usage:   "container rootfs",
			Aliases: []string{"cpu-quota"},
		}),
		altsrc.NewIntFlag(&cli.IntFlag{
			Name:    "Cgroups.cpu-period",
			Value:   0,
			Usage:   "Limit CPU CFS (Completely Fair Scheduler) period",
			Aliases: []string{"cpu-period"},
		}),
		altsrc.NewIntFlag(&cli.IntFlag{
			Name:    "Cgroups.cpu-rt-quota",
			Value:   0,
			Usage:   "How many time CPU will use in realtime scheduling (in usecs)",
			Aliases: []string{"cpu-rt-quota"},
		}),
		altsrc.NewIntFlag(&cli.IntFlag{
			Name:    "Cgroups.cpu-rt-period",
			Value:   0,
			Usage:   "Limit the CPU real-time period in microseconds",
			Aliases: []string{"cpu-rt-period"},
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    "Cgroups.cpuset-cpus",
			Value:   "",
			Usage:   "CPUs in which to allow execution (0-3, 0,1)",
			Aliases: []string{"cpuset-cpus"},
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    "Cgroups.cpuset-mems",
			Value:   "",
			Usage:   "MEMs in which to allow execution (0-3, 0,1)",
			Aliases: []string{"cpuset-mems"},
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    "Cgroups.kernel-memory",
			Value:   "20m",
			Usage:   "Kernel memory limit",
			Aliases: []string{"kernel-memory"},
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    "Cgroups.memory",
			Value:   "40M",
			Usage:   "Memory limit",
			Aliases: []string{"memory"},
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    "Cgroups.memory-swap",
			Value:   "80M",
			Usage:   "Total memory usage (memory + swap); set `-1` to enable unlimited swap",
			Aliases: []string{"memory-swap"},
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    "Cgroups.memory-reservation",
			Value:   "40M",
			Usage:   "Memory soft limit",
			Aliases: []string{"memory-reservation"},
		}),
		altsrc.NewIntFlag(&cli.IntFlag{
			Name:    "Cgroups.pids-limit",
			Value:   0,
			Usage:   "Tune container pids limit (set -1 for unlimited)",
			Aliases: []string{"pids-limit"},
		}),
	}
}
