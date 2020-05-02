package client

import (
	"context"

	rpcx "github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/protocol"
	"github.com/xhebox/chrootd/utils"
)

type rpcxClient struct {
	cli *rpcx.Client
}

func (c *rpcxClient) Call(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}) error {
	return c.cli.Call(ctx, servicePath, serviceMethod, args, reply)
}

func (c *rpcxClient) RemoteAddr() *utils.Addr {
	r := c.cli.Conn.RemoteAddr()
	return utils.NewAddr(r.Network(), r.String())
}

func (c *rpcxClient) Close() error {
	return c.cli.Close()
}

func newRpcxClient(network, addr string) (Client, error) {
	cli := rpcx.NewClient(rpcx.Option{SerializeType: protocol.MsgPack})
	err := cli.Connect(network, addr)
	return &rpcxClient{cli: cli}, err
}
