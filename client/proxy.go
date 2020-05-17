package client

import (
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"github.com/xhebox/chrootd/utils"
)

type Proxy struct {
	net  string
	svc  string
	meta map[string]string
	cli  Client
	con  *api.Client
}

func NewProxy(svc string, network string, cli interface{}, meta map[string]string) (*Proxy, error) {
	p := &Proxy{svc: svc, net: network, meta: meta}

	if cli == nil {
		return nil, errors.New("should provide a non-empty client")
	}

	switch v := cli.(type) {
	case Client:
		p.cli = v
	case *api.Client:
		p.con = v
	default:
		return nil, errors.New("cli must be a rpc client or a consul client")
	}

	return p, nil
}

func (m *Proxy) Call(id string, f func(Client, map[string]string) error) error {
	if m.cli != nil {
		return f(m.cli, m.meta)
	}

	ids := strings.SplitN(id, ",", 2)

	if m.con != nil {
		catalog := m.con.Catalog()
		svcs, _, err := catalog.Service(m.svc, ids[0], nil)
		if err != nil {
			return err
		}

		if len(svcs) == 0 {
			return errors.New("maybe the node holding this object is down")
		}

		svc := svcs[0]

		addr := utils.NewAddr(m.net, svc.ServiceAddress, svc.ServicePort)

		cli, err := NewClient(addr.Network(), addr.String())
		if err != nil {
			return err
		}
		defer cli.Close()

		return f(cli, svc.ServiceMeta)
	}

	return errors.New("internal error")
}

func (m *Proxy) Oneshot(id string, f func(Client) error) error {
	if m.cli != nil {
		return f(m.cli)
	}

	if m.con != nil {
		catalog := m.con.Catalog()

		svcs, _, err := catalog.Service(m.svc, id, nil)
		if err != nil {
			return nil
		}

		cnt := -1
		for _, svc := range svcs {
			addr := utils.NewAddr(m.net, svc.ServiceAddress, svc.ServicePort)

			cli, err := NewClient(addr.Network(), addr.String())
			if err != nil {
				continue
			}

			err = f(cli)

			cli.Close()
			if err == nil {
				cnt++
				break
			}
		}

		if cnt == -1 {
			return errors.New("all nodes failed")
		}
	}

	return nil
}

func (m *Proxy) Broadcast(f func(Client) error) error {
	if m.cli != nil {
		return f(m.cli)
	}

	if m.con != nil {
		catalog := m.con.Catalog()

		svcs, _, err := catalog.Service(m.svc, "", nil)
		if err != nil {
			return nil
		}

		cnt := -1

		errch := make(chan error, 1)

		for _, svc := range svcs {
			go func(svc *api.CatalogService) {
				addr := utils.NewAddr(m.net, svc.ServiceAddress, svc.ServicePort)

				cli, err := NewClient(addr.Network(), addr.String())
				if err != nil {
					return
				}

				err = f(cli)

				cli.Close()

				if err == nil {
					cnt++
				}

				errch <- err
			}(svc)
		}

		for range svcs {
			<-errch
		}

		if cnt == -1 {
			return errors.New("all nodes failed")
		}
	}

	return nil
}

func (m *Proxy) Close() error {
	if m.cli != nil {
		return m.cli.Close()
	}
	return nil
}
