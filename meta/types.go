package meta

import (
	"context"

	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/opencontainers/runtime-spec/specs-go"
)

const (
	PathMaskRead = iota
	PathMaskWrite
)

type PathMask struct {
	Mask int    `json:"mask"`
	Path string `json:"path"`
}

type Metainfo struct {
	Id             string               `json:"id"`
	Name           string               `json:"name"`
	Image          string               `json:"image"`
	ImageReference string               `json:"imagereference"`
	MaskPaths      []PathMask           `json:"maskPaths"`
	Mount          []specs.Mount        `json:"mount"`
	Hostname       string               `json:"hostname"`
	UidMapSize     uint32               `json:"uidMapSize"`
	GidMapSize     uint32               `json:"gidMapSize"`
	Resources      configs.Resources    `json:"cgroup"`
	Capabilities   configs.Capabilities `json:"capabilities"`
	Rlimits        []specs.POSIXRlimit  `json:"rlimits"`
	RootfsIds      []string             `json:"rootfsIds"`
}

type Manager interface {
	ID() (string, error)

	Create(spec *Metainfo) (string, error)
	Get(string) (*Metainfo, error)
	Update(*Metainfo) error
	Delete(string) error
	Query(string, func(*Metainfo) error) error

	ImageUnpack(context.Context, string) (string, error)
	ImageDelete(string, string) error
	ImageList(string, func(string) error) error
	ImageAvailable(context.Context, func(string, string, []string) error) error

	Close() error
}
