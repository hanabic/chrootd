package local

import (
	"path/filepath"
	"strings"

	"github.com/opencontainers/runc/libcontainer/configs"
	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"golang.org/x/sys/unix"
)

var mountMap = map[string]struct {
	clear bool
	flag  int
}{
	"acl":           {false, unix.MS_POSIXACL},
	"async":         {true, unix.MS_SYNCHRONOUS},
	"atime":         {true, unix.MS_NOATIME},
	"bind":          {false, unix.MS_BIND},
	"defaults":      {false, 0},
	"dev":           {true, unix.MS_NODEV},
	"diratime":      {true, unix.MS_NODIRATIME},
	"dirsync":       {false, unix.MS_DIRSYNC},
	"exec":          {true, unix.MS_NOEXEC},
	"iversion":      {false, unix.MS_I_VERSION},
	"lazytime":      {false, unix.MS_LAZYTIME},
	"loud":          {true, unix.MS_SILENT},
	"mand":          {false, unix.MS_MANDLOCK},
	"noacl":         {true, unix.MS_POSIXACL},
	"noatime":       {false, unix.MS_NOATIME},
	"nodev":         {false, unix.MS_NODEV},
	"nodiratime":    {false, unix.MS_NODIRATIME},
	"noexec":        {false, unix.MS_NOEXEC},
	"noiversion":    {true, unix.MS_I_VERSION},
	"nolazytime":    {true, unix.MS_LAZYTIME},
	"nomand":        {true, unix.MS_MANDLOCK},
	"norelatime":    {true, unix.MS_RELATIME},
	"nostrictatime": {true, unix.MS_STRICTATIME},
	"nosuid":        {false, unix.MS_NOSUID},
	"rbind":         {false, unix.MS_BIND | unix.MS_REC},
	"relatime":      {false, unix.MS_RELATIME},
	"remount":       {false, unix.MS_REMOUNT},
	"ro":            {false, unix.MS_RDONLY},
	"rw":            {true, unix.MS_RDONLY},
	"silent":        {false, unix.MS_SILENT},
	"strictatime":   {false, unix.MS_STRICTATIME},
	"suid":          {true, unix.MS_NOSUID},
	"sync":          {false, unix.MS_SYNCHRONOUS},
}

var mountPropagationMap = map[string]int{
	"private":     unix.MS_PRIVATE,
	"shared":      unix.MS_SHARED,
	"slave":       unix.MS_SLAVE,
	"unbindable":  unix.MS_UNBINDABLE,
	"rprivate":    unix.MS_PRIVATE | unix.MS_REC,
	"rshared":     unix.MS_SHARED | unix.MS_REC,
	"rslave":      unix.MS_SLAVE | unix.MS_REC,
	"runbindable": unix.MS_UNBINDABLE | unix.MS_REC,
}

func (m *CntrManager) spec2runcMounts(mounts []rspec.Mount) []*configs.Mount {
	res := []*configs.Mount{}
	for _, v := range mounts {
		pa := filepath.Clean(v.Destination)
		switch {
		case pa == "/etc/resolv.conf" && m.BinResolv:
			res = append(res, &configs.Mount{
				Source:      "/etc/resolv.conf",
				Destination: "/etc/resolv.conf",
				Flags:       unix.MS_BIND | unix.MS_REC | unix.MS_RDONLY,
			})
		case strings.HasPrefix(pa, "/proc"),
			strings.HasPrefix(pa, "/sys"),
			strings.HasPrefix(pa, "/dev"):
		default:
			flag := 0
			propagation := []int{}
			for _, o := range v.Options {
				if f, exists := mountMap[o]; exists && f.flag != 0 {
					if f.clear {
						flag &= ^f.flag
					} else {
						flag |= f.flag
					}
				} else if f, exists := mountPropagationMap[o]; exists && f != 0 {
					propagation = append(propagation, f)
				}
			}

			res = append(res, &configs.Mount{
				Source:           v.Source,
				Destination:      v.Destination,
				Device:           v.Type,
				Flags:            flag,
				PropagationFlags: propagation,
			})
		}
	}
	return res
}
