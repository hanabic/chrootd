package main

import (
	"github.com/docker/go-units"
	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/cntr"
	"github.com/xhebox/chrootd/utils"
	"golang.org/x/sys/unix"
)

func MetaFromCli(c *cli.Context) (cntr.Metainfo, error) {
	defaultMountFlags := unix.MS_NOEXEC | unix.MS_NOSUID | unix.MS_NODEV

	return cntr.Metainfo{
		Name:  c.String("name"),
		Rootfs: c.String("rootfs"),
		Capabilities: &configs.Capabilities{
			Bounding:    c.StringSlice("capabilities.bounding"),
			Effective:   c.StringSlice("capabilities.effective"),
			Inheritable: c.StringSlice("capabilities.inheritable"),
			Permitted:   c.StringSlice("capabilities.permitted"),
			Ambient:     c.StringSlice("capabilities.ambient"),
		},
		Cgroups: &configs.Cgroup{
			Name:   "Test",
			Parent: "system",
			Resources: &configs.Resources{
				MemorySwappiness:  nil,
				AllowAllDevices:   nil,
				AllowedDevices:    configs.DefaultAllowedDevices,
				BlkioWeight:       uint16(c.Uint64("cgroups.blkio-weight")),
				CpuShares:         c.Uint64("cgroups.cpu-shares"),
				CpuQuota:          c.Int64("cgroups.cpu-quota"),
				CpuPeriod:         c.Uint64("cgroups.cpu-period"),
				CpuRtRuntime:      c.Int64("cgroups.cpu-rt-quota"),
				CpuRtPeriod:       c.Uint64("cgroups.cpu-rt-period"),
				CpusetCpus:        c.String("cgroups.cpuset-cpus"),
				CpusetMems:        c.String("cgroups.cpuset-mems"),
				KernelMemory:      c.Int64("cgroups.kernel-memory"),
				Memory:            c.Int64("cgroups.memory"),
				MemorySwap:        c.Int64("cgroups.memory-swap"),
				MemoryReservation: c.Int64("cgroups.memory-reservation"),
				PidsLimit:         c.Int64("cgroups.pids-limit"),
			},
		},
		Namespaces: []configs.Namespace{
			{Type: configs.NEWUTS},
			{Type: configs.NEWIPC},
			{Type: configs.NEWPID},
			{Type: configs.NEWNET},
			{Type: configs.NEWNS},
			{Type: configs.NEWUSER},
		},
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
	}, nil
}

var (
	capFlags = []cli.Flag{
		&cli.StringSliceFlag{
			Name: "capabilities.bounding",
			Value: cli.NewStringSlice(
				"CAP_SETUID",
				"CAP_SETGID",
			),
			Aliases: []string{"cap_bou"},
		},
		&cli.StringSliceFlag{
			Name: "capabilities.effective",
			Value: cli.NewStringSlice(
				"CAP_SETUID",
				"CAP_SETGID",
			),
			Aliases: []string{"cap_eff"},
		},
		&cli.StringSliceFlag{
			Name: "capabilities.inheritable",
			Value: cli.NewStringSlice(
				"CAP_SETUID",
				"CAP_SETGID",
			),
			Aliases: []string{"cap_inh"},
		},
		&cli.StringSliceFlag{
			Name: "capabilities.permitted",
			Value: cli.NewStringSlice(
				"CAP_SETUID",
				"CAP_SETGID",
			),
			Aliases: []string{"cap_per"},
		},
		&cli.StringSliceFlag{
			Name: "capabilities.ambient",
			Value: cli.NewStringSlice(
				"CAP_SETUID",
				"CAP_SETGID",
			),
			Aliases: []string{"cap_amb"},
		},
	}

	cgroupFlags = []cli.Flag{
		&cli.Uint64Flag{
			Name:    "cgroups.blkio-weight",
			Value:   0,
			Usage:   "Block IO (relative weight), between 10 and 1000, or 0 to disable (default 0)",
			Aliases: []string{"blkio-weight"},
		},
		&cli.Uint64Flag{
			Name:    "cgroups.cpu-shares",
			Value:   0,
			Usage:   "CPU shares (relative weight vs. other containers",
			Aliases: []string{"cpu-shares"},
		},
		&cli.Int64Flag{
			Name:    "cgroups.cpu-quota",
			Value:   0,
			Usage:   "container rootfs",
			Aliases: []string{"cpu-quota"},
		},
		&cli.Uint64Flag{
			Name:    "cgroups.cpu-period",
			Value:   0,
			Usage:   "Limit CPU CFS (Completely Fair Scheduler) period",
			Aliases: []string{"cpu-period"},
		},
		&cli.Int64Flag{
			Name:    "cgroups.cpu-rt-quota",
			Value:   0,
			Usage:   "How many time CPU will use in realtime scheduling (in usecs)",
			Aliases: []string{"cpu-rt-quota"},
		},
		&cli.Uint64Flag{
			Name:    "cgroups.cpu-rt-period",
			Value:   0,
			Usage:   "Limit the CPU real-time period in microseconds",
			Aliases: []string{"cpu-rt-period"},
		},
		&cli.StringFlag{
			Name:    "cgroups.cpuset-cpus",
			Value:   "",
			Usage:   "CPUs in which to allow execution (0-3, 0,1)",
			Aliases: []string{"cpuset-cpus"},
		},
		&cli.StringFlag{
			Name:    "cgroups.cpuset-mems",
			Value:   "",
			Usage:   "MEMs in which to allow execution (0-3, 0,1)",
			Aliases: []string{"cpuset-mems"},
		},
		&utils.SizeFlag{
			Name:    "cgroups.kernel-memory",
			Value:   20 * units.MB,
			Usage:   "Kernel memory limit",
			Aliases: []string{"kernel-memory"},
		},
		&utils.SizeFlag{
			Name:    "cgroups.memory",
			Value:   40 * units.MB,
			Usage:   "Memory limit",
			Aliases: []string{"memory"},
		},
		&utils.SizeFlag{
			Name:    "cgroups.memory-swap",
			Value:   80 * units.MB,
			Usage:   "Total memory usage (memory + swap); set `-1` to enable unlimited swap",
			Aliases: []string{"memory-swap"},
		},
		&utils.SizeFlag{
			Name:    "cgroups.memory-reservation",
			Value:   40 * units.MB,
			Usage:   "Memory soft limit",
			Aliases: []string{"memory-reservation"},
		},
		&cli.Int64Flag{
			Name:    "cgroups.pids-limit",
			Value:   0,
			Usage:   "Tune container pids limit (set -1 for unlimited)",
			Aliases: []string{"pids-limit"},
		},
	}

	flags = utils.ConcatMultipleFlags(capFlags, cgroupFlags)
)
