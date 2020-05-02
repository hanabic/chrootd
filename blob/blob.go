package blob

import (
	"os"
	"path/filepath"
	"time"

	"github.com/bluele/gcache"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"github.com/tidwall/gjson"
	"github.com/xhebox/chrootd/store"
)

type Manager struct {
	path    string
	states  store.Store
	rtokens gcache.Cache
	wtokens gcache.Cache

	RTokenSize int
	WTokenSize int
}

func New(dir string, s store.Store, opts ...func(*Manager) error) (*Manager, error) {
	res := &Manager{
		path:       dir,
		states:     s,
		RTokenSize: 64,
		WTokenSize: 64,
	}
	for k := range opts {
		err := opts[k](res)
		if err != nil {
			return nil, err
		}
	}
	res.rtokens = gcache.New(res.RTokenSize).LRU().Build()
	res.wtokens = gcache.New(res.WTokenSize).LRU().Build()
	return res, nil
}

func (m *Manager) ReadToken(blobID string) (string, error) {
	ok, err := m.states.Has(blobID)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errors.New("no such blob")
	}

	token := ksuid.New().String()
	err = m.rtokens.SetWithExpire(token, blobID, time.Minute)
	return token, err
}

func (m *Manager) Read(token string) (*os.File, error) {
	v, err := m.rtokens.GetIFPresent(token)
	if err != nil {
		return nil, err
	}

	if !m.rtokens.Remove(token) {
		return nil, errors.New("can not remove token")
	}

	fid, ok := v.(string)
	if !ok {
		return nil, errors.New("maybe it is not a read token")
	}

	return os.Open(filepath.Join(m.path, fid))
}

func (m *Manager) WriteToken(meta string) (string, error) {
	if !gjson.Valid(meta) {
		return "", errors.New("invalid meta")
	}
	token := ksuid.New().String()
	err := m.wtokens.SetWithExpire(token, meta, time.Minute)
	return token, err
}

func (m *Manager) Write(token string) (*os.File, error) {
	v, err := m.wtokens.GetIFPresent(token)
	if err != nil {
		return nil, err
	}

	if !m.wtokens.Remove(token) {
		return nil, errors.New("can not remove token")
	}

	info, ok := v.(string)
	if !ok {
		return nil, errors.New("can not retrive meta")
	}

	err = m.states.Put(token, []byte(info))
	if err != nil {
		return nil, err
	}

	return os.OpenFile(filepath.Join(m.path, token), os.O_CREATE|os.O_RDWR, 0644)
}

func (m *Manager) Update(id string, meta string) error {
	if !gjson.Valid(meta) {
		return errors.New("invalid meta")
	}
	return m.states.Put(id, []byte(meta))
}

func (m *Manager) Delete(id string) error {
	os.Remove(filepath.Join(m.path, id))
	return m.states.Delete(id)
}

func (m *Manager) GetMeta(id string) (string, error) {
	it, err := m.states.Get(id)
	if err != nil {
		return "", err
	}
	return string(it), nil
}

type Blob struct {
	Id   string
	Meta string
}

func (m *Manager) List(query string) ([]Blob, error) {
	res := []Blob{}
	err := m.states.List("", func(k string, v []byte) error {
		if gjson.GetBytes(v, query).Type != gjson.Null {
			res = append(res, Blob{
				Id:   k,
				Meta: string(v),
			})
		}
		return nil
	})
	return res, err
}
