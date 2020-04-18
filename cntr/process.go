package cntr

import (
	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/configs"
)

type Proc struct {
	Args          []string
	Env           []string
	User          string
	Groups        []string
	Cwd           string
	ConsoleWidth  uint16
	ConsoleHeight uint16
	Capabilities  configs.Capabilities
	Rlimits       []configs.Rlimit
}

func (p *Proc) ToProc() *libcontainer.Process {
	return &libcontainer.Process{
		Args:             p.Args,
		Env:              p.Env,
		User:             p.User,
		AdditionalGroups: p.Groups,
		Cwd:              p.Cwd,
		ConsoleWidth:     p.ConsoleWidth,
		ConsoleHeight:    p.ConsoleHeight,
		Capabilities:     &p.Capabilities,
		Rlimits:          p.Rlimits,
	}
}
