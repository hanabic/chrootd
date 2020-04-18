package cntr

import (
	"github.com/docker/libkv/store"
)

type PeerDiscovery struct {
	peer string
}

func NewPeer(addr string) Discovery {
	return &PeerDiscovery{peer: addr}
}

func (r *PeerDiscovery) List() ([]*store.KVPair, error) {
	ret := []*store.KVPair{
		&store.KVPair{
			Key: r.peer,
		},
	}
	return ret, nil
}

func (r *PeerDiscovery) Addr() string {
	return r.peer
}

type MultipleDiscovery struct {
	srvs []string
}

func NewMultiple(addr string, others ...string) Discovery {
	if len(others) == 0 {
		return NewPeer(addr)
	} else {
		return &MultipleDiscovery{srvs: append(others, addr)}
	}
}

func (r *MultipleDiscovery) List() ([]*store.KVPair, error) {
	ret := []*store.KVPair{}
	for _, srv := range r.srvs {
		ret = append(ret, &store.KVPair{
			Key: srv,
		})
	}
	return ret, nil
}
