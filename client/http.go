package client

import (
	"context"
	"fmt"

	"github.com/ybbus/jsonrpc"
	"github.com/smallnest/rpcx/share"
)

type httpClient struct {
	cli  jsonrpc.RPCClient
}

func (c *httpClient) Call(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}) error {
	if reply == nil {
		reply = &struct{}{}
	}
	return c.cli.CallFor(reply, fmt.Sprintf("%s.%s", servicePath, serviceMethod), args, ctx.Value(share.ReqMetaDataKey))
}

func (c *httpClient) Close() error {
	return nil
}

func newHttpClient(addr string) (Client, error) {
	return &httpClient{cli: jsonrpc.NewClient(addr)}, nil
}
