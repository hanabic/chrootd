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
}

type Attacher interface {
	io.ReadWriteCloser
	CloseWrite() error
}

type Cntr interface {
	Meta() (*mtyp.Metainfo, error)
	Start(*Taskinfo) (string, error)
	Stop(string, bool) error
	StopAll(bool) error
	Wait() error
	Attach(string) (Attacher, error)
	List(func(string) error) error
}

type Manager interface {
	ID() (string, error)

	Create(*mtyp.Metainfo, string) (string, error)
	Get(string) (Cntr, error)
	Delete(string) error
	List(string, func(string, *mtyp.Metainfo) error) error

	Close() error
}
