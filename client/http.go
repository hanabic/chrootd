package client

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/smallnest/rpcx/codec"
	"github.com/smallnest/rpcx/server"
	"github.com/xhebox/chrootd/utils"
)

type httpClient struct {
	seq  uint64
	mu   sync.Mutex
	cli  *http.Client
	addr string
}

func (c *httpClient) Call(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}) error {
	cc := &codec.MsgpackCodec{}

	data, err := cc.Encode(args)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodConnect, c.addr, bytes.NewReader(data))
	if err != nil {
		return err
	}

	h := req.Header
	h.Set(server.XMessageID, fmt.Sprint(c.seq))
	c.mu.Lock()
	c.seq++
	c.mu.Unlock()
	h.Set(server.XMessageType, "0")
	h.Set(server.XSerializeType, "3")
	h.Set(server.XServicePath, servicePath)
	h.Set(server.XServiceMethod, serviceMethod)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	mb, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	return cc.Decode(mb, reply)
}

func (c *httpClient) RemoteAddr() *utils.Addr {
	return utils.NewAddr("http", c.addr)
}

func (c *httpClient) Close() error {
	return nil
}

func newHttpClient(addr string) (Client, error) {
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = fmt.Sprintf("http://%s", addr)
	}
	return &httpClient{seq: 1000, cli: &http.Client{}, addr: addr}, nil
}
