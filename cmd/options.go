package main

import (
	"strings"

	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/urfave/cli/v2"
	ctyp "github.com/xhebox/chrootd/cntr"
	mtyp "github.com/xhebox/chrootd/meta"
	"github.com/xhebox/chrootd/utils"
)

// TODO: read from file
func MetaFromCli(c *cli.Context) (*mtyp.Metainfo, error) {
	img := strings.SplitN(c.String("image"), ":", 2)
	if len(img) < 2 {
		img = []string{c.String("image"), "latest"}
	}

	res := &mtyp.Metainfo{
		Id:             c.String("id"),
		Name:           c.String("name"),
		Image:          img[0],
		ImageReference: img[1],
		Capabilities: configs.Capabilities{
			Bounding:    c.StringSlice("capabilities.bounding"),
			Effective:   c.StringSlice("capabilities.effective"),
			Inheritable: c.StringSlice("capabilities.inheritable"),
			Permitted:   c.StringSlice("capabilities.permitted"),
			Ambient:     c.StringSlice("capabilities.ambient"),
		},
		Resources: configs.Resources{
			BlkioWeight:       uint16(c.Uint64("resources.blkio-weight")),
			CpuShares:         c.Uint64("resources.cpu-shares"),
			CpuQuota:          c.Int64("resources.cpu-quota"),
			CpuPeriod:         c.Uint64("resources.cpu-period"),
			CpuRtRuntime:      c.Int64("resources.cpu-rt-quota"),
			CpuRtPeriod:       c.Uint64("resources.cpu-rt-period"),
			CpusetCpus:        c.String("resources.cpuset-cpus"),
			CpusetMems:        c.String("resources.cpuset-mems"),
			KernelMemory:      c.Int64("resources.kernel-memory"),
			Memory:            c.Int64("resources.memory"),
			MemorySwap:        c.Int64("resources.memory-swap"),
			MemoryReservation: c.Int64("resources.memory-reservation"),
			PidsLimit:         c.Int64("resources.pids-limit"),
		},
	}

	return res, nil
}

func TaskFromCli(c *cli.Context) (*ctyp.Taskinfo, error) {
	res := &ctyp.Taskinfo{
		Args: c.Args().Slice(),
		Env:  c.StringSlice("env"),
		Capabilities: configs.Capabilities{
			Bounding:    c.StringSlice("capabilities.bounding"),
			Effective:   c.StringSlice("capabilities.effective"),
			Inheritable: c.StringSlice("capabilities.inheritable"),
			Permitted:   c.StringSlice("capabilities.permitted"),
			Ambient:     c.StringSlice("capabilities.ambient"),
		},
	}
	return res, nil
}

var (
	taskFlags = []cli.Flag{
		&cli.StringSliceFlag{
			Name:  "env",
			Usage: "additional env",
		},
		&cli.StringFlag{
			Name:  "file",
			Usage: "read config from file",
		},
	}

	capFlags = []cli.Flag{
		&cli.StringSliceFlag{
			Name:    "capabilities.bounding",
			Aliases: []string{"cap_bou"},
		},
		&cli.StringSliceFlag{
			Name:    "capabilities.effective",
			Aliases: []string{"cap_eff"},
		},
		&cli.StringSliceFlag{
			Name:    "capabilities.inheritable",
			Aliases: []string{"cap_inh"},
		},
		&cli.StringSliceFlag{
			Name:    "capabilities.permitted",
			Aliases: []string{"cap_per"},
		},
		&cli.StringSliceFlag{
			Name:    "capabilities.ambient",
			Aliases: []string{"cap_amb"},
		},
	}

	resourceFlags = []cli.Flag{
		&cli.Uint64Flag{
			Name:    "resources.blkio-weight",
			Usage:   "Block IO (relative weight), between 10 and 1000, or 0 to disable (default 0)",
			Aliases: []string{"blkio-weight"},
		},
		&cli.Uint64Flag{
			Name:    "resources.cpu-shares",
			Usage:   "CPU shares (relative weight vs. other containers)",
			Aliases: []string{"cpu-shares"},
		},
		&cli.Int64Flag{
			Name:    "resources.cpu-quota",
			Usage:   "CPU CFS period to be used for hardcapping (in usecs). 0 to use system default",
			Aliases: []string{"cpu-quota"},
		},
		&cli.Uint64Flag{
			Name:    "resources.cpu-period",
			Usage:   "Limit CPU CFS (Completely Fair Scheduler) period",
			Aliases: []string{"cpu-period"},
		},
		&cli.Int64Flag{
			Name:    "resources.cpu-rt-quota",
			Usage:   "How many time CPU will use in realtime scheduling (in usecs)",
			Aliases: []string{"cpu-rt-quota"},
		},
		&cli.Uint64Flag{
			Name:    "resources.cpu-rt-period",
			Usage:   "Limit the CPU real-time period in microseconds",
			Aliases: []string{"cpu-rt-period"},
		},
		&cli.StringFlag{
			Name:    "resources.cpuset-cpus",
			Usage:   "CPUs in which to allow execution (0-3, 0,1)",
			Aliases: []string{"cpuset-cpus"},
		},
		&cli.StringFlag{
			Name:    "resources.cpuset-mems",
			Usage:   "MEMs in which to allow execution (0-3, 0,1)",
			Aliases: []string{"cpuset-mems"},
		},
		&utils.SizeFlag{
			Name:    "resources.kernel-memory",
			Usage:   "Kernel memory limit",
			Aliases: []string{"kernel-memory"},
		},
		&utils.SizeFlag{
			Name:    "resources.memory",
			Usage:   "Memory limit",
			Aliases: []string{"memory"},
		},
		&utils.SizeFlag{
			Name:    "resources.memory-swap",
			Usage:   "Total memory usage (memory + swap); set `-1` to enable unlimited swap",
			Aliases: []string{"memory-swap"},
		},
		&utils.SizeFlag{
			Name:    "resources.memory-reservation",
			Usage:   "Memory soft limit",
			Aliases: []string{"memory-reservation"},
		},
		&cli.Int64Flag{
			Name:    "resources.pids-limit",
			Usage:   "Tune container pids limit (set -1 for unlimited)",
			Aliases: []string{"pids-limit"},
		},
	}

	metaFlags = []cli.Flag{
		&cli.StringFlag{
			Name:  "name",
			Value: "hello",
			Usage: "containere name",
		},
		&cli.StringFlag{
			Name:  "id",
			Usage: "suggest which node this metadata will be created on",
		},
		&cli.StringFlag{
			Name:  "image",
			Usage: "image reference",
		},
		&cli.StringFlag{
			Name:  "file",
			Usage: "read config from file",
		},
	}
)
