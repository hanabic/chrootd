package local

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/openSUSE/umoci/oci/cas/dir"
	"github.com/openSUSE/umoci/oci/casext"
	"github.com/openSUSE/umoci/oci/layer"
	ispec "github.com/opencontainers/image-spec/specs-go/v1"
	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"github.com/tidwall/gjson"
	. "github.com/xhebox/chrootd/meta"
	"github.com/xhebox/chrootd/store"
	"github.com/xhebox/chrootd/utils"
)

type MetaManager struct {
	id         string
	runPath    string
	imagePath  string
	rootfsPath string
	metas      store.Store

	Rootless     bool
	DefaultImage string
}

func NewMetaManager(path, image string, s store.Store, opts ...func(*MetaManager) error) (Manager, error) {
	mgr := &MetaManager{
		imagePath:    image,
		runPath:      filepath.Join(path, "meta"),
		rootfsPath:   filepath.Join(path, "rootfs"),
		Rootless:     true,
		DefaultImage: "alpine",
	}

	for k := range opts {
		err := opts[k](mgr)
		if err != nil {
			return nil, err
		}
	}

	if !utils.PathExist(mgr.imagePath) {
		return nil, errors.Errorf("image path does not exist or no permission: %s", mgr.imagePath)
	}

	err := os.MkdirAll(mgr.runPath, 0755)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(mgr.rootfsPath, 0755)
	if err != nil {
		return nil, err
	}

	mgrid, err := store.LoadOrStore(s, "id", ksuid.New().String())
	if err != nil {
		return nil, err
	}
	mgr.id = string(mgrid)

	mgrmetas, err := store.NewWrapStore("meta", s)
	if err != nil {
		return nil, err
	}
	mgr.metas = mgrmetas

	return mgr, nil
}

func (m *MetaManager) specValid(spec *Metainfo) *Metainfo {
	if spec.UidMapSize == 0 {
		spec.UidMapSize = 1
	}
	if spec.GidMapSize == 0 {
		spec.GidMapSize = 1
	}
	if spec.Hostname == "" {
		spec.Hostname = "chrootd"
	}
	if len(spec.RootfsIds) > 0 {
		spec.RootfsIds = nil
	}
	if spec.ImageReference == "" {
		spec.ImageReference = "latest"
	}
	if spec.Image == "" {
		spec.Image = m.DefaultImage
	}
	return spec
}

func (m *MetaManager) getMeta(id string) (uint64, *Metainfo, error) {
	idx, it, err := m.metas.Get(id)
	if err != nil {
		return 0, nil, err
	}

	res := &Metainfo{}
	err = json.Unmarshal(it, res)
	return idx, res, err
}

func (m *MetaManager) putMeta(idx uint64, id string, meta *Metainfo) error {
	meta.Id = id

	mb, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	return m.metas.Put(id, idx, mb)
}

func (m *MetaManager) ID() (string, error) {
	return m.id, nil
}

func (m *MetaManager) Create(spec *Metainfo) (string, error) {
	spec = m.specValid(spec)
	newid, err := m.metas.NextSequence()
	if err != nil {
		return "", err
	}
	err = m.putMeta(0, utils.ComposeID(m.id, fmt.Sprint(newid)), spec)
	return spec.Id, err
}

func (m *MetaManager) Get(id string) (*Metainfo, error) {
	_, res, err := m.getMeta(id)
	return res, err
}

func (m *MetaManager) Update(spec *Metainfo) error {
	idx, meta, err := m.getMeta(spec.Id)
	if err != nil {
		return err
	}

	if len(meta.RootfsIds) > 0 {
		return errors.New("metadata is not writable while its rootfs is unpacked")
	}

	spec = m.specValid(spec)
	spec.RootfsIds = meta.RootfsIds

	return m.putMeta(idx, spec.Id, spec)
}

func (m *MetaManager) Delete(mid string) error {
	idx, meta, err := m.getMeta(mid)
	if err != nil {
		return err
	}

	if len(meta.RootfsIds) > 0 {
		return errors.New("metadata is not writable while its rootfs is unpacked")
	}

	return m.metas.Delete(mid, idx)
}

func (m *MetaManager) Query(query string, f func(v *Metainfo) error) error {
	return m.metas.List("", func(k string, idx uint64, v []byte) error {
		if query == "" || gjson.GetBytes(v, query).Type != gjson.Null {
			res := &Metainfo{}
			err := json.Unmarshal(v, res)
			if err != nil {
				return err
			}
			return f(res)
		}
		return nil
	})
}

func (m *MetaManager) ImageUnpack(ctx context.Context, metaid string) (string, error) {
	idx, meta, err := m.getMeta(metaid)
	if err != nil {
		return "", err
	}

	readonly := false
	for _, mask := range meta.MaskPaths {
		if path.Clean(mask.Path) == "/" && mask.Mask&PathMaskWrite != 0 {
			readonly = true
			break
		}
	}

	if len(meta.RootfsIds) > 0 && readonly {
		return meta.RootfsIds[0], nil
	}

	ce, err := dir.Open(filepath.Join(m.imagePath, meta.Image))
	if err != nil {
		return "", err
	}
	cext := casext.NewEngine(ce)
	defer cext.Close()

	desc, err := cext.ResolveReference(ctx, meta.ImageReference)
	if err != nil {
		return "", err
	}

	if len(desc) != 1 {
		return "", errors.New("non-exist or ambiguous reference")
	}

	manifestBlob, err := cext.FromDescriptor(ctx, desc[0].Descriptor())
	if err != nil {
		return "", err
	}
	defer manifestBlob.Close()

	if manifestBlob.Descriptor.MediaType != ispec.MediaTypeImageManifest {
		return "", errors.Errorf("except a manifest file: %s", manifestBlob.Descriptor.MediaType)
	}

	manifest, ok := manifestBlob.Data.(ispec.Manifest)
	if !ok {
		return "", errors.Errorf("should be here, internal corruption")
	}

	id := ksuid.New().String()

	path := filepath.Join(m.rootfsPath, id)

	defer func() {
		if err != nil {
			os.RemoveAll(path)
		}
	}()

	opt := &layer.MapOptions{
		Rootless: m.Rootless,
		UIDMappings: []rspec.LinuxIDMapping{
			rspec.LinuxIDMapping{
				ContainerID: 0,
				HostID:      uint32(os.Geteuid()),
				Size:        meta.UidMapSize,
			},
		},
		GIDMappings: []rspec.LinuxIDMapping{
			rspec.LinuxIDMapping{
				ContainerID: 0,
				HostID:      uint32(os.Getegid()),
				Size:        meta.GidMapSize,
			},
		},
	}
	err = layer.UnpackRootfs(ctx, cext.Engine, path, manifest, opt)
	if err != nil {
		return "", err
	}

	meta.RootfsIds = append(meta.RootfsIds, id)

	return id, m.putMeta(idx, metaid, meta)
}

func (m *MetaManager) ImageDelete(metaid, rootid string) error {
	idx, meta, err := m.getMeta(metaid)
	if err != nil {
		return err
	}

	i := -1
	for j := range meta.RootfsIds {
		if meta.RootfsIds[j] == rootid {
			i = j
			break
		}
	}
	if i >= 0 {
		if i < len(meta.RootfsIds) {
			copy(meta.RootfsIds[i:], meta.RootfsIds[i+1:])
		}
		meta.RootfsIds = meta.RootfsIds[:len(meta.RootfsIds)-1]

		err := os.RemoveAll(filepath.Join(m.runPath, rootid))
		if err != nil {
			return err
		}
	}

	return m.putMeta(idx, metaid, meta)
}

func (m *MetaManager) ImageList(metaid string, f func(string) error) error {
	_, meta, err := m.getMeta(metaid)
	if err != nil {
		return err
	}

	for k := range meta.RootfsIds {
		if err := f(meta.RootfsIds[k]); err != nil {
			return err
		}
	}

	return nil
}

func (m *MetaManager) ImageAvailable(ctx context.Context, f func(string, string, []string) error) error {
	imgs, err := os.Open(m.imagePath)
	if err != nil {
		return err
	}
	defer imgs.Close()

	images, err := imgs.Readdirnames(-1)
	if err != nil {
		return err
	}

	for _, image := range images {
		ce, err := dir.Open(filepath.Join(m.imagePath, image))
		if err != nil {
			return err
		}
		cext := casext.NewEngine(ce)

		refs, err := cext.ListReferences(ctx)
		if err != nil {
			cext.Close()
			return err
		}

		err = f(m.id, image, refs)
		if err != nil {
			cext.Close()
			return err
		}

		cext.Close()
	}

	return nil
}

func (m *MetaManager) Close() error {
	return nil
}
