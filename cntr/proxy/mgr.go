package proxy

import (
	"context"

	"github.com/xhebox/chrootd/client"
	ctyp "github.com/xhebox/chrootd/cntr"
	mtyp "github.com/xhebox/chrootd/meta"
	"github.com/xhebox/chrootd/utils"
)

type CntrProxy struct {
	*client.Proxy
	svc     string
	Network string
	Context context.Context
}

func NewCntrProxy(svcname string, cli interface{}, attachAddr *utils.Addr, opts ...func(*CntrProxy) error) (ctyp.Manager, error) {
	mgr := &CntrProxy{svc: svcname, Network: "tcp", Context: context.Background()}

	for i := range opts {
		if err := opts[i](mgr); err != nil {
			return nil, err
		}
	}

	meta := map[string]string{}
	if attachAddr != nil {
		meta["attachNetwork"] = attachAddr.Network()
		meta["attach"] = attachAddr.String()
	}
	pro, err := client.NewProxy(svcname, mgr.Network, cli, meta)
	if err != nil {
		return nil, err
	}
	mgr.Proxy = pro

	return mgr, nil
}

func (m *CntrProxy) ID() (string, error) {
	return "", nil
}

func (m *CntrProxy) Create(meta *mtyp.Metainfo, rid string) (string, error) {
	res := ""
	return res, m.Call(meta.Id, func(cli client.Client, svc map[string]string) error {
		return cli.Call(m.Context, m.svc, "Create", &CreateReq{
			Meta:   meta,
			Rootfs: rid,
		}, &res)
	})
}

func (m *CntrProxy) Delete(cid string) error {
	return m.Call(cid, func(cli client.Client, svc map[string]string) error {
		return cli.Call(m.Context, m.svc, "Delete", cid, nil)
	})
}

func (m *CntrProxy) List(node string, f func(string, *mtyp.Metainfo) error) error {
	nid, query, _ := utils.DecomposeID(node)
	if nid != "" {
		return m.Call(nid, func(cli client.Client, svc map[string]string) error {
			res := []CntrListRes{}

			err := cli.Call(m.Context, m.svc, "List", query, &res)
			if err != nil {
				return err
			}

			for _, v := range res {
				if err := f(v.Id, v.Meta); err != nil {
					return err
				}
			}

			return nil
		})
	}

	return m.Broadcast(func(cli client.Client) error {
		res := []CntrListRes{}

		err := cli.Call(m.Context, m.svc, "List", node, &res)
		if err != nil {
			return err
		}

		for _, v := range res {
			if err := f(v.Id, v.Meta); err != nil {
				return err
			}
		}

		return nil
	})
}

func (m *CntrProxy) Get(id string) (ctyp.Cntr, error) {
	return &cntr{cid: id, CntrProxy: m}, nil
}
