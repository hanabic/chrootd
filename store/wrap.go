package store

import (
	"fmt"
	"path"
	"strings"

	"github.com/pkg/errors"
)

func NewWrapStore(prefix string, under Store) (Store, error) {
	if under == nil {
		return nil, errors.New("nil under storage")
	}
	switch v := under.(type) {
	case *boltStore:
		return v.appendPrefix(prefix)
	default:
		return &wrapStore{
			Store:  v,
			prefix: fmt.Sprintf("%s/", prefix),
		}, nil
	}
}

type wrapStore struct {
	Store
	prefix string
}

func (w *wrapStore) Get(i string) (uint64, []byte, error) {
	return w.Store.Get(path.Join(w.prefix, i))
}

func (w *wrapStore) Put(k string, idx uint64, v []byte) error {
	return w.Store.Put(path.Join(w.prefix, k), idx, v)
}

func (w *wrapStore) Delete(i string, idx uint64) error {
	return w.Store.Delete(path.Join(w.prefix, i), idx)
}

func (w *wrapStore) List(k string, f func(string, uint64, []byte) error) error {
	return w.Store.List(path.Join(w.prefix, k), func(k string, idx uint64, v []byte) error {
		return f(strings.TrimPrefix(k, w.prefix), idx, v)
	})
}

func (w *wrapStore) Has(k string) (bool, error) {
	return w.Store.Has(path.Join(w.prefix, k))
}

func (w *wrapStore) NextSequence() (uint64, error) {
	return w.Store.NextSequence()
}
