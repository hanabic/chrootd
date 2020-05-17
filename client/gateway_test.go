package client

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"

	"github.com/smallnest/rpcx/server"
	"github.com/xhebox/chrootd/utils"
	"github.com/ybbus/jsonrpc"
)

type TestGateway struct {
	rpcx *server.Server
	gw   *Gateway
	srv  http.Server
	cli  jsonrpc.RPCClient
}

func NewTestGateway() (*TestGateway, error) {
	res := &TestGateway{}

	rpcAddr := utils.NewAddrFree()

	lnRPC, err := net.Listen(rpcAddr.Network(), rpcAddr.String())
	if err != nil {
		return nil, err
	}

	srv := server.NewServer(func(m *server.Server) {
		m.DisableHTTPGateway = true
		m.DisableJSONRPC = true
	})
	res.rpcx = srv

	err = srv.RegisterFunctionName("t", "t", func(ctx context.Context, req *string, res *struct{}) error {
		return nil
	}, "")
	if err != nil {
		return nil, err
	}

	go res.rpcx.ServeListener(rpcAddr.Network(), lnRPC)

	addr := utils.NewAddrFree()

	lnHTTP, err := net.Listen(addr.Network(), addr.String())
	if err != nil {
		return nil, err
	}

	rcl, err := NewGateway(rpcAddr.Network(), rpcAddr.String())
	if err != nil {
		return nil, err
	}
	res.gw = rcl
	res.srv.Handler = rcl

	go res.srv.Serve(lnHTTP)

	res.cli = jsonrpc.NewClient(fmt.Sprintf("http://%s/", addr.String()))

	return res, nil
}

func (m *TestGateway) Close() error {
	m.gw.Close()
	m.srv.Shutdown(context.Background())
	return m.rpcx.Shutdown(context.Background())
}

func TestGatewayMeta(t *testing.T) {
	mgr, err := NewTestGateway()
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	reply := &struct{}{}
	err = mgr.cli.CallFor(reply, "t.t", map[string]string{"ss": "Hh"}, "tt")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGatewayNormal(t *testing.T) {
	mgr, err := NewTestGateway()
	if err != nil {
		t.Fatal(err)
	}
	defer mgr.Close()

	reply := &struct{}{}
	err = mgr.cli.CallFor(reply, "t.t", "tt")
	if err != nil {
		t.Fatal(err)
	}
}
