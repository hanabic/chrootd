package main

import (
	"errors"

	"github.com/segmentio/ksuid"
	"github.com/xhebox/chrootd/api"
	bolt "go.etcd.io/bbolt"
)

type cntrPool struct {
	storage *bolt.DB
}

func newCntrPool(path string) (*cntrPool, error) {
	db, err := bolt.Open(path, 0644, &bolt.Options{})
	if err != nil {
		return nil, err
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("containerPools"))
		return err
	}); err != nil {
		return nil, err
	}

	return &cntrPool{
		storage: db,
	}, nil
}

func (p *cntrPool) Add(meta []byte) ([]byte, error) {
	newid := ksuid.New()

	if err := p.storage.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucket(newid.Bytes())
		if err != nil {
			return err
		}

		m, err := api.NewMetaFromBytes(meta)
		if err != nil {
			return err
		}
		m.Id = newid.Bytes()

		cfg, err := m.ToBytes()
		if err != nil {
			return err
		}

		if err := bucket.Put([]byte("metadata"), cfg); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return newid.Bytes(), nil
}

func (p *cntrPool) UpdateMeta(id []byte, meta []byte) ([]byte, error) {
	var cfg []byte

	if err := p.storage.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(id)
		if bucket == nil {
			return errors.New("bucket does not exist")
		}

		metadata := bucket.Get([]byte("metadata"))

		m1, err := api.NewMetaFromBytes(metadata)
		if err != nil {
			return err
		}

		m2, err := api.NewMetaFromBytes(meta)
		if err != nil {
			return err
		}

		m1.Merge(m2)

		cfg, err = m1.ToBytes()
		if err != nil {
			return err
		}

		if err := bucket.Put([]byte("metadata"), cfg); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (p *cntrPool) Del(id []byte) error {
	if err := p.storage.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(id)
		if bucket == nil {
			return errors.New("bucket does not exist")
		}

		busy := bucket.Get([]byte("busy"))
		if len(busy) != 0 {
			return errors.New("busy")
		}

		if err := tx.DeleteBucket(id); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (p *cntrPool) StartMeta(id []byte) (*api.Metainfo, error) {
	var metadata []byte

	if err := p.storage.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(id)
		if bucket == nil {
			return errors.New("bucket does not exist")
		}

		cnt := bucket.Get([]byte("busy"))
		if len(cnt) == 0 {
			cnt = []byte{1}
		} else {
			cnt[0] += 1
		}

		if err := bucket.Put([]byte("busy"), cnt); err != nil {
			return err
		}

		metadata = bucket.Get([]byte("metadata"))

		return nil
	}); err != nil {
		return nil, err
	}

	return api.NewMetaFromBytes(metadata)
}

func (p *cntrPool) StopMeta(id []byte) error {
	return p.storage.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(id)
		if bucket == nil {
			return errors.New("bucket does not exist")
		}

		cnt := bucket.Get([]byte("busy"))
		if len(cnt) > 0 && cnt[0] > 0 {
			cnt[0] -= 1
		}

		if err := bucket.Put([]byte("busy"), cnt); err != nil {
			return err
		}

		return nil
	})
}

func (p *cntrPool) ForEach(f func([]byte, *bolt.Bucket) error) error {
	if err := p.storage.View(func(tx *bolt.Tx) error {
		return tx.ForEach(f)
	}); err != nil {
		return err
	}

	return nil
}

func (p *cntrPool) Close() {
	p.storage.Close()
}
