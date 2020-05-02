package store

type StoreList interface {
	List(string, func(string, []byte) error) error
}

type StoreGet interface {
	Get(string) ([]byte, error)
}

type StoreClose interface {
	Close() error
}

type StorePut interface {
	Put(string, []byte) error
}

type StoreDelete interface {
	Delete(string) error
}

type StoreHas interface {
	Has(string) (bool, error)
}

type Store interface {
	StoreList
	StoreGet
	StorePut
	StoreDelete
	StoreHas
	StoreClose
}
