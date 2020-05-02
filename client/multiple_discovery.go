package client

/*
import (
	"github.com/pkg/errors"
	"github.com/xhebox/chrootd/store"
)

type MultipleDiscovery struct {
	srvs []string
}

func NewMultiple(addrs ...string) (Discovery, error) {
	if len(addrs) == 0 {
		return nil, errors.Errorf("no address")
	} else if len(addrs) == 1 {
		return NewPeer(addrs[0]), nil
	} else {
		return &MultipleDiscovery{srvs: addrs}, nil
	}
}

func (r *MultipleDiscovery) List(string) ([]*store.Item, error) {
	ret := []*store.Item{}
	for _, srv := range r.srvs {
		ret = append(ret, &store.Item{
			Val: []byte(srv),
		})
	}
	return ret, nil
}

func (r *MultipleDiscovery) Close() error {
	return nil
}
*/
