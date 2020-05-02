package store

import (
	"path"
)

type wrapStore struct {
	Store
	prefix string
}

func NewWrapStore(prefix string, under Store) Store {
	if under == nil {
		return nil
	}
	return &wrapStore{
		Store:  under,
		prefix: prefix,
	}
}

func (w *wrapStore) Get(i string) ([]byte, error) {
	return w.Store.Get(path.Join(w.prefix, i))
}

func (w *wrapStore) Put(k string, v []byte) error {
	return w.Store.Put(path.Join(w.prefix, k), v)
}

func (w *wrapStore) Delete(i string) error {
	return w.Store.Delete(path.Join(w.prefix, i))
}

func (w *wrapStore) List(k string, f func(string, []byte) error) error {
	return w.Store.List(path.Join(w.prefix, k), f)
}

func (w *wrapStore) Has(k string) (bool, error) {
	return w.Store.Has(path.Join(w.prefix, k))
}
