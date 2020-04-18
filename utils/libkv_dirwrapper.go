package utils

import (
	"path"

	"github.com/docker/libkv/store"
)

type KVDir struct {
	prefix string
	under  store.Store
}

func NewKVDir(prefix string, under store.Store) *KVDir {
	return &KVDir{
		prefix: prefix,
		under:  under,
	}
}

func (k *KVDir) Put(key string, value []byte, options *store.WriteOptions) error {
	return k.under.Put(path.Join(k.prefix, key), value, options)
}

func (k *KVDir) Get(key string) (*store.KVPair, error) {
	return k.under.Get(path.Join(k.prefix, key))
}

func (k *KVDir) Delete(key string) error {
	return k.under.Delete(path.Join(k.prefix, key))
}

func (k *KVDir) Exists(key string) (bool, error) {
	return k.under.Exists(path.Join(k.prefix, key))
}

func (k *KVDir) Watch(key string, stopCh <-chan struct{}) (<-chan *store.KVPair, error) {
	return k.under.Watch(path.Join(k.prefix, key), stopCh)
}

func (k *KVDir) WatchTree(directory string, stopCh <-chan struct{}) (<-chan []*store.KVPair, error) {
	return k.under.WatchTree(path.Join(k.prefix, directory), stopCh)
}

func (k *KVDir) NewLock(key string, options *store.LockOptions) (store.Locker, error) {
	return k.under.NewLock(path.Join(k.prefix, key), options)
}

func (k *KVDir) List(directory string) ([]*store.KVPair, error) {
	return k.under.List(path.Join(k.prefix, directory))
}

func (k *KVDir) DeleteTree(directory string) error {
	return k.under.DeleteTree(path.Join(k.prefix, directory))
}

func (k *KVDir) AtomicPut(key string, value []byte, previous *store.KVPair, options *store.WriteOptions) (bool, *store.KVPair, error) {
	return k.under.AtomicPut(path.Join(k.prefix, key), value, previous, options)
}

func (k *KVDir) AtomicDelete(key string, previous *store.KVPair) (bool, error) {
	return k.under.AtomicDelete(path.Join(k.prefix, key), previous)
}

func (k *KVDir) Close() {
	k.under.Close()
}
