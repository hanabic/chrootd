package cntr

import (
	"io"

	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/opencontainers/runtime-spec/specs-go"
	mtyp "github.com/xhebox/chrootd/meta"
)

type Taskinfo struct {
	Args         []string             `json:"args"`
	Env          []string             `json:"env"`
	Capabilities configs.Capabilities `json:"capabilities"`
	Rlimits      []specs.POSIXRlimit  `json:"rlimits"`
	TermHeight   uint16               `json:"term_height"`
	TermWidth    uint16               `json:"term_width"`
}

type Cntrinfo struct {
	Id     string
	Rootfs string
	Tags   []string
	Meta   *mtyp.Metainfo
}

type Attacher interface {
	io.ReadWriteCloser
	CloseWrite() error
}

type Cntr interface {
	Meta() (*Cntrinfo, error)
	Start(*Taskinfo) (string, error)
	Stop(string, bool) error
	StopAll(bool) error
	Wait() error
	Attach(string) (Attacher, error)
	List(func(string) error) error
}

type Manager interface {
	ID() (string, error)

	Create(*Cntrinfo) (string, error)
	Get(string) (Cntr, error)
	Delete(string) error
	List(string, func(*Cntrinfo) error) error

	Close() error
}
