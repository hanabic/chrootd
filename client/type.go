package client

import (
	"context"

	"github.com/xhebox/chrootd/store"
	"github.com/xhebox/chrootd/utils"
)

type Discovery interface {
	store.StoreList
	store.StoreClose
}

type Registry interface {
	store.Store
}

type Client interface {
	 Call(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}) error
	 RemoteAddr() *utils.Addr
	 Close() error
}

type ClientPool struct {
	discovery Discovery
}

func NewClient(network, addr string) (Client, error) {
	if network == "http" {
		return newHttpClient(addr)
	} else {
		return newRpcxClient(network, addr)
	}
}
