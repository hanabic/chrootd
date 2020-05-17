package client

import (
	"context"
	"time"

	rpcx "github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/protocol"
)

type rpcxClient struct {
	cli *rpcx.Client
}

func (c *rpcxClient) Call(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}) error {
	if reply == nil {
		reply = &struct{}{}
	}
	return c.cli.Call(ctx, servicePath, serviceMethod, args, reply)
}

func (c *rpcxClient) Close() error {
	if c.cli != nil {
		c.cli.Close()
	}
	return nil
}

func newRpcxClient(network, addr string) (Client, error) {
	cli := rpcx.NewClient(rpcx.Option{SerializeType: protocol.MsgPack, Heartbeat: true, HeartbeatInterval: time.Second})
	err := cli.Connect(network, addr)
	if err != nil {
		return nil, err
	}
	return &rpcxClient{cli: cli}, nil
}
