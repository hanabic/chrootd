package store

type Store interface {
	Has(string) (bool, error)
	Get(string) (uint64, []byte, error)
	List(string, func(string, uint64, []byte) error) error
	Delete(string, uint64) error
	Put(string, uint64, []byte) error
	NextSequence() (uint64, error)
	Close() error
}

func LoadOrStore(s Store, k, v string) (string, error) {
	_, idb, err := s.Get(k)
	if err != nil || len(idb) == 0 {
		err = s.Put(k, 0, []byte(v))
		return v, err
	}
	return string(idb), nil
}
