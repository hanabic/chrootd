package store

import (
	"encoding/binary"
	"strings"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

type boltStore struct {
	db     *bolt.DB
	bucket string
	prefix []string
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

func (s *boltStore) appendPrefix(k string) (*boltStore, error) {
	res := &boltStore{db: s.db, bucket: s.bucket, prefix: append(s.prefix, k)}

	err := res.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(res.bucket))
		if bkt == nil {
			return errors.New("no such bucket")
		}

		var err error
		for _, b := range res.prefix {
			bkt, err = bkt.CreateBucketIfNotExists([]byte(b))
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *boltStore) List(prefix string, f func(string, uint64, []byte) error) error {
	return s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(s.bucket))
		if bkt == nil {
			return errors.New("no such bucket")
		}

		for _, b := range s.prefix {
			bkt = bkt.Bucket([]byte(b))
			if bkt == nil {
				return errors.New("no such bucket")
			}
		}

		return bkt.ForEach(func(k, v []byte) error {
			if len(v) < 8 {
				return nil
			}

			idx := binary.BigEndian.Uint64(v[:8])

			kk := string(k)
			if strings.HasPrefix(kk, prefix) {
				return f(kk, idx, v[8:])
			}

			return nil
		})
	})
}

func (s *boltStore) Get(k string) (uint64, []byte, error) {
	res := []byte{}

	err := s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(s.bucket))
		if bkt == nil {
			return errors.New("no such bucket")
		}

		for _, b := range s.prefix {
			bkt = bkt.Bucket([]byte(b))
			if bkt == nil {
				return errors.New("no such bucket")
			}
		}

		res = bkt.Get([]byte(k))
		return nil
	})
	if err != nil {
		return 0, nil, err
	}
	if len(res) < 8 {
		return 0, nil, errors.New("no such key")
	}

	idx := binary.BigEndian.Uint64(res[:8])
	return idx, res[8:], nil
}

func (s *boltStore) Put(k string, idx uint64, v []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(s.bucket))
		if bkt == nil {
			return errors.New("no such bucket")
		}

		for _, b := range s.prefix {
			bkt = bkt.Bucket([]byte(b))
			if bkt == nil {
				return errors.New("no such bucket")
			}
		}

		old := bkt.Get([]byte(k))
		if old != nil {
			if len(old) < 8 {
				return errors.New("no such key")
			}

			oidx := binary.BigEndian.Uint64(old[:8])
			if oidx != idx {
				return errors.New("modified, can not atomically done")
			}
		}

		dbval := make([]byte, 8+len(v))
		binary.BigEndian.PutUint64(dbval, idx+1)
		copy(dbval[8:], v)
		return bkt.Put([]byte(k), dbval)
	})
}

func (s *boltStore) Delete(k string, idx uint64) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(s.bucket))
		if bkt == nil {
			return errors.New("no such bucket")
		}

		for _, b := range s.prefix {
			bkt = bkt.Bucket([]byte(b))
			if bkt == nil {
				return errors.New("no such bucket")
			}
		}

		old := bkt.Get([]byte(k))
		if len(old) < 8 {
			return errors.New("no such key")
		}

		oidx := binary.BigEndian.Uint64(old[:8])
		if oidx != idx {
			return errors.New("modified, can not atomically done")
		}

		return bkt.Delete([]byte(k))
	})
}

func (s *boltStore) Has(k string) (bool, error) {
	_, _, err := s.Get(k)
	return err == nil, err
}

func (s *boltStore) NextSequence() (uint64, error) {
	res := uint64(0)
	return res, s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(s.bucket))
		if bkt == nil {
			return errors.New("no such bucket")
		}

		for _, b := range s.prefix {
			bkt = bkt.Bucket([]byte(b))
			if bkt == nil {
				return errors.New("no such bucket")
			}
		}

		var err error
		res, err = bkt.NextSequence()
		return err
	})
}

func (s *boltStore) Close() error {
	return s.db.Close()
}
