package main

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/urfave/cli/v2"
	ctyp "github.com/xhebox/chrootd/cntr"
	mtyp "github.com/xhebox/chrootd/meta"
	"github.com/xhebox/chrootd/utils"
)

func capFromCli(res *configs.Capabilities, c *cli.Context) {
	if c.IsSet("capabilities.bounding") {
		res.Bounding = c.StringSlice("capabilities.bounding")
	}

	if c.IsSet("capabilities.effective") {
		res.Effective = c.StringSlice("capabilities.effective")
	}

	if c.IsSet("capabilities.inheritable") {
		res.Inheritable = c.StringSlice("capabilities.inheritable")
	}

	if c.IsSet("capabilities.permitted") {
		res.Permitted = c.StringSlice("capabilities.permitted")
	}

	if c.IsSet("capabilities.ambient") {
		res.Ambient = c.StringSlice("capabilities.ambient")
	}
}

func MetaFromCli(c *cli.Context) (*mtyp.Metainfo, error) {
	res := &mtyp.Metainfo{}

	if tmpl := c.String("file"); tmpl != "" {
		tb, err := ioutil.ReadFile(tmpl)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(tb, res)
		if err != nil {
			return nil, err
		}
	}

	if c.IsSet("image") {
		img := strings.SplitN(c.String("image"), ":", 2)
		if len(img) < 2 {
			img = []string{c.String("image"), "latest"}
		}
		res.Image = img[0]
		res.ImageReference = img[1]
	}

	if c.IsSet("id") {
		res.Id = c.String("id")
	}

	if c.IsSet("name") {
		res.Name = c.String("name")
	}

	capFromCli(&res.Capabilities, c)

	if c.IsSet("resources.blkio-weight") {
		res.Resources.BlkioWeight = uint16(c.Uint64("resources.blkio-weight"))
	}

	if c.IsSet("resources.cpu-shares") {
		res.Resources.CpuShares = c.Uint64("resources.cpu-shares")
	}

	if c.IsSet("resources.cpu-quota") {
		res.Resources.CpuQuota = c.Int64("resources.cpu-quota")
	}

	if c.IsSet("resources.cpu-period") {
		res.Resources.CpuPeriod = c.Uint64("resources.cpu-period")
	}

	if c.IsSet("resources.cpu-rt-quota") {
		res.Resources.CpuRtRuntime = c.Int64("resources.cpu-rt-quota")
	}

	if c.IsSet("resources.cpu-rt-period") {
		res.Resources.CpuRtPeriod = c.Uint64("resources.cpu-rt-period")
	}

	if c.IsSet("resources.cpuset-cpus") {
		res.Resources.CpusetCpus = c.String("resources.cpuset-cpus")
	}

	if c.IsSet("resources.cpuset-mems") {
		res.Resources.CpusetMems = c.String("resources.cpuset-mems")
	}

	if c.IsSet("resources.kernel-memory") {
		res.Resources.KernelMemory = c.Int64("resources.kernel-memory")
	}
	if c.IsSet("resources.memory") {
		res.Resources.Memory = c.Int64("resources.memory")
	}

	if c.IsSet("resources.memory-swap") {
		res.Resources.MemorySwap = c.Int64("resources.memory-swap")
	}

	if c.IsSet("resources.memory-reservation") {
		res.Resources.MemoryReservation = c.Int64("resources.memory-reservation")
	}

	if c.IsSet("resources.pids-limit") {
		res.Resources.PidsLimit = c.Int64("resources.pids-limit")
	}

	return res, nil
}

func TaskFromCli(c *cli.Context) (*ctyp.Taskinfo, error) {
	res := &ctyp.Taskinfo{}

	if tmpl := c.String("file"); tmpl != "" {
		tb, err := ioutil.ReadFile(tmpl)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(tb, res)
		if err != nil {
			return nil, err
		}
	}

	if c.Args().Len() > 0 {
		res.Args = c.Args().Slice()
	}

	if c.IsSet("env") {
		res.Env = c.StringSlice("env")
	}

	capFromCli(&res.Capabilities, c)

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
