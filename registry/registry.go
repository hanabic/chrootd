package registry

import (
	"path"

	"github.com/docker/libkv/store"
)

type Discovery interface {
	List() ([]*store.KVPair, error)
	Close() error
}

type Registry interface {
	Discovery

	Get(string) (*store.KVPair, error)
	Put(string, []byte) error
	Delete(string) error
}

type StoreRegistry struct {
	store.Store
}

func NewStoreRegistry(s store.Store) Registry {
	return &StoreRegistry{Store: s}
}

func (s *StoreRegistry) List() ([]*store.KVPair, error) {
	return s.Store.List("")
}

func (s *StoreRegistry) Put(k string, v []byte) error {
	return s.Store.Put(k, v, &store.WriteOptions{})
}

func (s *StoreRegistry) Close() error {
	s.Store.Close()
	return nil
}

type WrapRegistry struct {
	Registry
	prefix string
}

func NewWrapRegistry(prefix string, under Registry) Registry {
	if under == nil {
		return nil
	}
	return &WrapRegistry{
		Registry: under,
		prefix:   prefix,
	}
}

func (w *WrapRegistry) Get(i string) (*store.KVPair, error) {
	return w.Registry.Get(path.Join(w.prefix, i))
}

func (w *WrapRegistry) Put(k string, v []byte) error {
	return w.Registry.Put(path.Join(w.prefix, k), v)
}

func (w *WrapRegistry) Delete(i string) error {
	return w.Registry.Delete(path.Join(w.prefix, i))
}
