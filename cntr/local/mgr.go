package local

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/imdario/mergo"
	"github.com/opencontainers/runc/libcontainer"
	"github.com/opencontainers/runc/libcontainer/cgroups"
	"github.com/opencontainers/runc/libcontainer/cgroups/systemd"
	"github.com/opencontainers/runc/libcontainer/configs"
	_ "github.com/opencontainers/runc/libcontainer/nsenter"
	"github.com/opencontainers/runc/libcontainer/specconv"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"github.com/tidwall/gjson"
	"github.com/urfave/cli/v2"
	. "github.com/xhebox/chrootd/cntr"
	mtyp "github.com/xhebox/chrootd/meta"
	"github.com/xhebox/chrootd/store"
	"github.com/xhebox/chrootd/utils"
	"golang.org/x/sys/unix"
)

func InitLibcontainer() {
	runtime.GOMAXPROCS(1)
	runtime.LockOSThread()

	factory, _ := libcontainer.New("")
	if err := factory.StartInitialization(); err != nil {
		os.Exit(1)
	}
}

var (
	InitFlag = &cli.Command{
		Name:   "___init",
		Hidden: true,
		Action: func(c *cli.Context) error {
			InitLibcontainer()
			return nil
		},
	}
)

type CntrManager struct {
	id          string
	imagePath   string
	rootfsPath  string
	factoryPath string
	factory     libcontainer.Factory

	states store.Store
	cntrs  map[string]*cntr
	rwmux  sync.RWMutex

	Rootless  bool
	BinResolv bool
}

func NewCntrManager(path, image string, s store.Store, opts ...func(*CntrManager) error) (*CntrManager, error) {
	mgr := &CntrManager{
		imagePath:   image,
		factoryPath: filepath.Join(path, "factory"),
		rootfsPath:  filepath.Join(path, "rootfs"),
		cntrs:       make(map[string]*cntr),
		Rootless:    true,
		BinResolv:   true,
	}
	for _, f := range opts {
		err := f(mgr)
		if err != nil {
			return nil, err
		}
	}

	if !utils.PathExist(mgr.imagePath) {
		return nil, errors.New("image path does not exist or no permission")
	}

	err := os.MkdirAll(mgr.rootfsPath, 0755)
	if err != nil {
		return nil, err
	}

	mgrid, err := store.LoadOrStore(s, "id", ksuid.New().String())
	if err != nil {
		return nil, err
	}
	mgr.id = string(mgrid)

	mgrstates, err := store.NewWrapStore("cntr", s)
	if err != nil {
		return nil, err
	}
	mgr.states = mgrstates

	cgroupMgr := libcontainer.Cgroupfs
	if systemd.IsRunningSystemd() {
		cgroupMgr = libcontainer.SystemdCgroups
		if cgroups.IsCgroup2UnifiedMode() && mgr.Rootless {
			cgroupMgr = libcontainer.RootlessSystemdCgroups
		}
	} else if mgr.Rootless {
		cgroupMgr = libcontainer.RootlessCgroupfs
	}

	uidPath, err := exec.LookPath("newuidmap")
	if err != nil {
		uidPath = "/bin/newuidmap"
	}

	gidPath, err := exec.LookPath("newgidmap")
	if err != nil {
		gidPath = "/bin/newgidmap"
	}

	err = os.RemoveAll(mgr.factoryPath)
	if err != nil {
		return nil, err
	}

	mgr.factory, err = libcontainer.New(mgr.factoryPath,
		cgroupMgr,
		libcontainer.InitArgs(os.Args[0], "___init"),
		// without suid/guid or corressponding caps, extern mapping tools are needed to run rootless(with correct configuration)
		libcontainer.NewuidmapPath(uidPath),
		libcontainer.NewgidmapPath(gidPath),
	)
	if err != nil {
		return nil, err
	}

	return mgr, nil
}

func (m *CntrManager) getCntr(id string) (*cntr, error) {
	m.rwmux.Lock()
	defer m.rwmux.Unlock()

	c, ok := m.cntrs[id]
	if !ok {
		return nil, errors.New("container does not exist")
	}
	return c, nil
}

func (m *CntrManager) ID() (string, error) {
	return m.id, nil
}

func (m *CntrManager) Create(meta *mtyp.Metainfo, rootfs string) (string, error) {
	cfg := &configs.Config{
		Rootfs: filepath.Join(m.rootfsPath, rootfs),
		Cgroups: &configs.Cgroup{
			Name:      "container",
			Resources: &meta.Resources,
		},
		Namespaces: configs.Namespaces{
			{Type: configs.NEWUTS},
			{Type: configs.NEWIPC},
			{Type: configs.NEWPID},
			{Type: configs.NEWNS},
			{Type: configs.NEWUSER},
		},
		MaskPaths: []string{
			"/proc/acpi",
			"/proc/asound",
			"/proc/kcore",
			"/proc/keys",
			"/proc/latency_stats",
			"/proc/timer_list",
			"/proc/timer_stats",
			"/proc/sched_debug",
			"/proc/scsi",
			"/sys/firmware",
		},
		ReadonlyPaths: []string{
			"/proc/bus",
			"/proc/fs",
			"/proc/irq",
			"/proc/sys",
			"/proc/sysrq-trigger",
		},
		Devices: specconv.AllowedDevices,
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
		},
		UidMappings: []configs.IDMap{
			configs.IDMap{
				ContainerID: 0,
				HostID:      os.Geteuid(),
				Size:        int(meta.UidMapSize),
			},
		},
		GidMappings: []configs.IDMap{
			configs.IDMap{
				ContainerID: 0,
				HostID:      os.Getegid(),
				Size:        int(meta.GidMapSize),
			},
		},
		Hostname:        meta.Hostname,
		RootlessCgroups: m.Rootless,
		RootlessEUID:    m.Rootless,
		Hooks: &configs.Hooks{
			Prestart: []configs.Hook{},
		},
	}

	if !m.Rootless {
		cfg.Namespaces = append(cfg.Namespaces, configs.Namespace{Type: configs.NEWNET})
		cfg.Mounts = append(cfg.Mounts, &configs.Mount{
				Source:      "sysfs",
				Destination: "/sys",
				Device:      "sysfs",
				Flags:       unix.MS_NOEXEC | unix.MS_NOSUID | unix.MS_NODEV | unix.MS_RDONLY,
			})
	} else {
		cfg.Mounts = append(cfg.Mounts, &configs.Mount{
				Source:      "/sys",
				Destination: "/sys",
				Flags:       unix.MS_BIND | unix.MS_REC | unix.MS_RDONLY,
			})
	}

	for _, v := range meta.MaskPaths {
		if v.Mask&mtyp.PathMaskRead != 0 {
			if filepath.Clean(v.Path) == "/" {
				cfg.Readonlyfs = true
			} else {
				cfg.ReadonlyPaths = append(cfg.ReadonlyPaths, v.Path)
			}
		}
		if v.Mask&mtyp.PathMaskWrite != 0 {
			cfg.MaskPaths = append(cfg.MaskPaths, v.Path)
		}
	}

	cfg.Rlimits = append(cfg.Rlimits, spec2runcRlimits(meta.Rlimits)...)
	cfg.Mounts = append(cfg.Mounts, m.spec2runcMounts(meta.Mount)...)

	err := mergo.Merge(cfg.Cgroups.Resources, meta.Resources)
	if err != nil {
		return "", err
	}

	newid, err := m.states.NextSequence()
	if err != nil {
		return "", err
	}

	id := utils.ComposeID(m.id, fmt.Sprint(newid))

	c, err := m.factory.Create(id, cfg)
	if err != nil {
		return "", err
	}

	m.rwmux.Lock()
	m.cntrs[id] = newCntr(c, meta)
	m.rwmux.Unlock()

	return id, nil
}

func (m *CntrManager) Get(id string) (Cntr, error) {
	return m.getCntr(id)
}

func (m *CntrManager) Delete(id string) error {
	cntr, err := m.getCntr(id)
	if err != nil {
		return err
	}

	err = cntr.Destroy()
	if err != nil {
		return err
	}

	m.rwmux.Lock()
	_, ok := m.cntrs[id]
	if ok {
		delete(m.cntrs, id)
	}
	m.rwmux.Unlock()

	return nil
}

func (m *CntrManager) List(id string, f func(k string, meta *mtyp.Metainfo) error) error {
	m.rwmux.RLock()
	defer m.rwmux.RUnlock()

	for k, cntr := range m.cntrs {
		mb, _ := json.Marshal(cntr.meta)
		if gjson.ValidBytes(mb) && gjson.GetBytes(mb, id).Type != gjson.Null {
			if err := f(k, cntr.meta); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *CntrManager) Close() error {
	m.rwmux.RLock()
	defer m.rwmux.RUnlock()

	for _, cntr := range m.cntrs {
		cntr.Destroy()
	}

	return nil
}
