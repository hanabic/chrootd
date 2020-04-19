package cntr

import (
	"encoding/json"
	"golang.org/x/sys/unix"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/imdario/mergo"
	"github.com/opencontainers/runc/libcontainer/configs"
)

var (
	DefaultSecureConfig = &Metainfo{
		Cgroups: &configs.Cgroup{
			Name:   "container",
			Parent: "system",
			Resources: &configs.Resources{
				AllowedDevices: configs.DefaultAllowedDevices,
			},
		},
		Namespaces: configs.Namespaces{
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
				Flags:       unix.MS_NOEXEC | unix.MS_NOSUID | unix.MS_NODEV,
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
				Flags:       unix.MS_NOEXEC | unix.MS_NOSUID | unix.MS_NODEV | unix.MS_RDONLY,
			},
		},
	}
)

type Metainfo struct {
	Name          string                `json:"name"`
	Rootfs        string                `json:"rootfs"`
	Hostname      string                `json:"hostname"`
	Readonlyfs    bool                  `json:"readonlyfs"`
	ReadonlyPaths []string              `json:"readonly_paths,omitempty"`
	Mounts        []*configs.Mount      `json:"mounts,omitempty"`
	Devices       []*configs.Device     `json:"devices,omitempty"`
	Namespaces    configs.Namespaces    `json:"namespaces,omitempty"`
	Capabilities  *configs.Capabilities `json:"capabilities,omitempty"`
	Networks      []*configs.Network    `json:"networks,omitempty"`
	Routes        []*configs.Route      `json:"routes,omitempty"`
	Cgroups       *configs.Cgroup       `json:"cgroups,omitempty"`
	Rlimits       []configs.Rlimit      `json:"rlimits,omitempty"`
	Seccomp       *configs.Seccomp      `json:"seccomp,omitempty"`
	Sysctl        map[string]string     `json:"sysctl,omitempty"`
	UidMappings   []configs.IDMap       `json:"uid_mappings,omitempty"`
	GidMappings   []configs.IDMap       `json:"gid_mappings,omitempty"`
}

func NewMetaFromBytes(bytes []byte) (*Metainfo, error) {
	r := &Metainfo{}
	if err := json.Unmarshal(bytes, r); err != nil {
		return nil, err
	}
	return r, nil
}

func NewMetaFromFile(path string) (*Metainfo, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	r := &Metainfo{}
	if err := json.Unmarshal(bytes, r); err != nil {
		return nil, err
	}
	return r, nil
}

func NewMetaFromConfig(cfg configs.Config) (*Metainfo, error) {
	return &Metainfo{
		Rootfs:        cfg.Rootfs,
		Hostname:      cfg.Hostname,
		Readonlyfs:    cfg.Readonlyfs,
		ReadonlyPaths: cfg.ReadonlyPaths,
		Mounts:        cfg.Mounts,
		Devices:       cfg.Devices,
		Namespaces:    cfg.Namespaces,
		Capabilities:  cfg.Capabilities,
		Networks:      cfg.Networks,
		Routes:        cfg.Routes,
		Cgroups:       cfg.Cgroups,
		Rlimits:       cfg.Rlimits,
		Seccomp:       cfg.Seccomp,
		Sysctl:        cfg.Sysctl,
		UidMappings:   cfg.UidMappings,
		GidMappings:   cfg.GidMappings,
	}, nil
}

func (m *Metainfo) Merge(o *Metainfo) error {
	return mergo.Merge(m, o, mergo.WithOverride, mergo.WithAppendSlice)
}

func (m Metainfo) ToBytes() ([]byte, error) {
	r, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (m Metainfo) String() string {
	strb, _ := json.MarshalIndent(m, "", "\t")
	return string(strb)
}

func (m Metainfo) ToConfig() configs.Config {
	return configs.Config{
		Rootfs:        m.Rootfs,
		Hostname:      m.Hostname,
		Readonlyfs:    m.Readonlyfs,
		ReadonlyPaths: m.ReadonlyPaths,
		Mounts:        m.Mounts,
		Devices:       m.Devices,
		Namespaces:    m.Namespaces,
		Capabilities:  m.Capabilities,
		Networks:      m.Networks,
		Routes:        m.Routes,
		Cgroups:       m.Cgroups,
		Rlimits:       m.Rlimits,
		Seccomp:       m.Seccomp,
		Sysctl:        m.Sysctl,
		UidMappings:   m.UidMappings,
		GidMappings:   m.GidMappings,
	}
}

func (m *Metainfo) idMap(mappings []configs.IDMap, src, dst, size int) []configs.IDMap {
	for k := range mappings {
		if mappings[k].ContainerID == src {
			return mappings
		}
	}

	return append(mappings, configs.IDMap{
		ContainerID: src,
		HostID:      dst,
		Size:        size,
	})
}

func (m *Metainfo) UIDMap(src, dst, size int) {
	if m.Namespaces.Contains(configs.NEWUSER) {
		m.UidMappings = m.idMap(m.UidMappings, src, dst, size)
	}
}

func (m *Metainfo) GIDMap(src, dst, size int) {
	if m.Namespaces.Contains(configs.NEWUSER) {
		m.GidMappings = m.idMap(m.GidMappings, src, dst, size)
	}
}

func (m *Metainfo) Default() error {
	err := m.Merge(DefaultSecureConfig)
	if err != nil {
		return err
	}

	m.UIDMap(0, os.Getuid(), 1)
	m.GIDMap(0, os.Getgid(), 1)

	if filepath.Clean(m.Rootfs) == "/" {
		m.Readonlyfs = true
	}

	return nil
}
