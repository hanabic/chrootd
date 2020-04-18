package cntr

import (
	"context"
	"fmt"
	"sync"

	"github.com/xhebox/chrootd/utils"
	"github.com/smallnest/rpcx/client"
)

type Client struct {
	servicePath string
	singlePeer string
	option      client.Option

	cntrs    Registry
	services Discovery

	cachedClient   map[string]*client.Client
	cachedClientMu sync.Mutex
}

func NewClient(servicePath string, srv Discovery, cntrs Registry, option client.Option) (*Client, error) {
	r := &Client{
		servicePath:  servicePath,
		services:     srv,
		cntrs:        cntrs,
		cachedClient: make(map[string]*client.Client),
		option:       option,
	}
	if srv == nil {
		return nil, fmt.Errorf("nil service discovery")
	}
	switch k := srv.(type) {
	case *PeerDiscovery:
		r.singlePeer = k.Addr()
	default:
		if cntrs == nil {
			return nil, fmt.Errorf("need container resolver")
		}
	}
	return r, nil
}

func (c *Client) getCachedClient(addr string) (*client.Client, error) {
	c.cachedClientMu.Lock()
	defer c.cachedClientMu.Unlock()

	cli, ok := c.cachedClient[addr]
	if ok {
		return cli, nil
	}

	na := utils.NewAddrFromString(addr)

	cli = client.NewClient(c.option)

	err := cli.Connect(na.Network(), na.Addr())
	if err != nil {
		return nil, err
	}

	c.cachedClient[addr] = cli

	return cli, nil
}

func (c *Client) getClient(id string) (*client.Client, error) {
	if len(c.singlePeer) != 0 {
		_srv, _ := c.services.List()
		return c.getCachedClient(string(_srv[0].Value))
	}

	cntr, err := c.cntrs.Get(id)
	if err != nil {
		return nil, err
	}

	cli, err := c.getCachedClient(string(cntr.Value))
	if err != nil {
		return nil, err
	}

	return cli, nil
}

func (c *Client) List(ctx context.Context, req *ListReq, res *ListRes) error {
	srvs, err := c.services.List()
	if err != nil {
		return err
	}

	for _, srv := range srvs {
		cli, err := c.getCachedClient(string(srv.Value))
		if err != nil {
			return err
		}

		r := &ListRes{}

		err = cli.Call(ctx, c.servicePath, "List", req, r)
		if err != nil {
			return err
		}

		res.CntrIds = append(res.CntrIds, r.CntrIds...)
	}

	for _, cntr := range res.CntrIds {
		err = c.cntrs.Put(cntr.Id, []byte(cntr.Addr))
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) Create(ctx context.Context, req *CreateReq, res *CreateRes) (e error) {
	srvs, err := c.services.List()
	if err != nil {
		return err
	}

	for _, srv := range srvs {
		cli, err := c.getCachedClient(string(srv.Key))
		if err != nil {
			return err
		}

		e = cli.Call(ctx, c.servicePath, "Create", req, res)
		if e == nil {
			e = c.cntrs.Put(res.Id, []byte(srv.Key))
			return
		}
	}

	return
}

func (c *Client) Config(ctx context.Context, req *ConfigReq, res *ConfigRes) error {
	cli, err := c.getClient(req.Id)
	if err != nil {
		return err
	}

	err = cli.Call(ctx, c.servicePath, "Config", req, res)
	if err == nil {
		return err
	}

	return nil
}

func (c *Client) Delete(ctx context.Context, req *DeleteReq, res *DeleteRes) error {
	cli, err := c.getClient(req.Id)
	if err != nil {
		return err
	}

	err = cli.Call(ctx, c.servicePath, "Delete", req, res)
	if err == nil {
		return err
	}

	err = c.cntrs.Delete(req.Id)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Exec(ctx context.Context, req *ExecReq, res *ExecRes) error {
	cli, err := c.getClient(req.Id)
	if err != nil {
		return err
	}

	err = cli.Call(ctx, c.servicePath, "Exec", req, res)
	if err == nil {
		return err
	}

	return nil
}

func (c *Client) Pause(ctx context.Context, req *PauseReq, res *PauseRes) error {
	cli, err := c.getClient(req.Id)
	if err != nil {
		return err
	}

	err = cli.Call(ctx, c.servicePath, "Pause", req, res)
	if err == nil {
		return err
	}

	return nil
}

func (c *Client) Resume(ctx context.Context, req *ResumeReq, res *ResumeRes) error {
	cli, err := c.getClient(req.Id)
	if err != nil {
		return err
	}

	err = cli.Call(ctx, c.servicePath, "Resume", req, res)
	if err == nil {
		return err
	}

	return nil
}

func (c *Client) Start(ctx context.Context, req *StartReq, res *StartRes) error {
	cli, err := c.getClient(req.Id)
	if err != nil {
		return err
	}

	err = cli.Call(ctx, c.servicePath, "Start", req, res)
	if err == nil {
		return err
	}

	return nil
}

func (c *Client) Stop(ctx context.Context, req *StopReq, res *StopRes) error {
	cli, err := c.getClient(req.Id)
	if err != nil {
		return err
	}

	err = cli.Call(ctx, c.servicePath, "Stop", req, res)
	if err == nil {
		return err
	}

	return nil
}

func (c *Client) Status(ctx context.Context, req *StatusReq, res *StatusRes) error {
	cli, err := c.getClient(req.Id)
	if err != nil {
		return err
	}

	err = cli.Call(ctx, c.servicePath, "Status", req, res)
	if err == nil {
		return err
	}

	return nil
}

func (c *Client) Close() (e error) {
	c.cachedClientMu.Lock()
	defer c.cachedClientMu.Unlock()

	for _, cli := range c.cachedClient {
		e = cli.Close()
	}

	return
}
