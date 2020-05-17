package client

import (
	"context"
)

type Client interface {
	Call(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}) error
	Close() error
}

func NewClient(network, addr string) (Client, error) {
	if network == "http" {
		return newHttpClient(addr)
	} else {
		return newRpcxClient(network, addr)
	}
}
