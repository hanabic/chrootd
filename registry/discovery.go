package registry

import (
	"fmt"

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
			Value: []byte(r.peer),
		},
	}
	return ret, nil
}

func (r *PeerDiscovery) Addr() string {
	return r.peer
}

func (r *PeerDiscovery) Close() error {
	return nil
}

type MultipleDiscovery struct {
	srvs []string
}

func NewMultiple(addrs ...string) (Discovery, error) {
	if len(addrs) == 0 {
		return nil, fmt.Errorf("no address")
	} else if len(addrs) == 1 {
		return NewPeer(addrs[0]), nil
	} else {
		return &MultipleDiscovery{srvs: addrs}, nil
	}
}

func (r *MultipleDiscovery) List() ([]*store.KVPair, error) {
	ret := []*store.KVPair{}
	for _, srv := range r.srvs {
		ret = append(ret, &store.KVPair{
			Value: []byte(srv),
		})
	}
	return ret, nil
}

func (r *MultipleDiscovery) Close() error {
	return nil
}
