package cntr

import (
	"github.com/docker/libkv/store"
)

type Discovery interface {
	List() ([]*store.KVPair, error)
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
