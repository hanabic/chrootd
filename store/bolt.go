package store

import (
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

type boltStore struct {
	db     *bolt.DB
	bucket string
}

func NewBolt(path string, bucket string) (Store, error) {
	db, err := bolt.Open(path, 0644, &bolt.Options{})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		return err
	})
	if err != nil {
		db.Close()
		return nil, err
	}

	return &boltStore{db: db, bucket: bucket}, nil
}

func (s *boltStore) List(k string, f func(string, []byte) error) error {
	return s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(s.bucket))
		if bkt == nil {
			return errors.New("no such bucket")
		}

		return bkt.ForEach(func(k, v []byte) error {
			return f(string(k), v)
		})
	})
}

func (s *boltStore) Get(k string) ([]byte, error) {
	res := []byte{}

	err := s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(s.bucket))
		if bkt == nil {
			return errors.New("no such bucket")
		}

		res = bkt.Get([]byte(k))
		return nil
	})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, errors.New("no such key")
	}

	return res, nil
}

func (s *boltStore) Put(k string, v []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(s.bucket))
		if bkt == nil {
			return errors.New("no such bucket")
		}

		return bkt.Put([]byte(k), v)
	})
}

func (s *boltStore) Delete(k string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(s.bucket))
		if bkt == nil {
			return errors.New("no such bucket")
		}

		return bkt.Delete([]byte(k))
	})
}

func (s *boltStore) Has(k string) (bool, error) {
	_, err := s.Get(k)
	return err == nil, err
}

func (s *boltStore) Close() error {
	return s.db.Close()
}
