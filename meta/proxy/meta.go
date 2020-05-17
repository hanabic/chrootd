package proxy

import (
	"context"

	"github.com/xhebox/chrootd/client"
	mtyp "github.com/xhebox/chrootd/meta"
)

type MetaProxy struct {
	*client.Proxy
	svc     string
	Network string
	Context context.Context
}

func NewMetaProxy(svcname string, cli interface{}, opts ...func(*MetaProxy) error) (mtyp.Manager, error) {
	mgr := &MetaProxy{svc: svcname, Context: context.Background(), Network: "tcp"}

	for i := range opts {
		if err := opts[i](mgr); err != nil {
			return nil, err
		}
	}

	pro, err := client.NewProxy(svcname, mgr.Network, cli, nil)
	if err != nil {
		return nil, err
	}
	mgr.Proxy = pro

	return mgr, nil
}

func (m *MetaProxy) ID() (string, error) {
	return "", nil
}

func (m *MetaProxy) Create(meta *mtyp.Metainfo) (string, error) {
	res := ""
	return res, m.Oneshot(meta.Id, func(cli client.Client) error {
		return cli.Call(m.Context, m.svc, "Create", meta, &res)
	})
}

func (m *MetaProxy) Get(id string) (*mtyp.Metainfo, error) {
	res := &mtyp.Metainfo{}
	return res, m.Call(id, func(cli client.Client, svc map[string]string) error {
		return cli.Call(m.Context, m.svc, "Get", id, res)
	})
}

func (m *MetaProxy) Update(meta *mtyp.Metainfo) error {
	return m.Call(meta.Id, func(cli client.Client, svc map[string]string) error {
		return cli.Call(m.Context, m.svc, "Update", meta, nil)
	})
}

func (m *MetaProxy) Delete(id string) error {
	return m.Call(id, func(cli client.Client, svc map[string]string) error {
		return cli.Call(m.Context, m.svc, "Delete", id, nil)
	})
}

func (m *MetaProxy) Query(query string, f func(*mtyp.Metainfo) error) error {
	return m.Broadcast(func(cli client.Client) error {
		res := []*mtyp.Metainfo{}

		err := cli.Call(m.Context, m.svc, "Query", query, &res)
		if err != nil {
			return err
		}

		for _, meta := range res {
			if err := f(meta); err != nil {
				return err
			}
		}

		return nil
	})
}

func (m *MetaProxy) ImageUnpack(ctx context.Context, mid string) (string, error) {
	res := ""
	return res, m.Call(mid, func(cli client.Client, svc map[string]string) error {
		return cli.Call(ctx, m.svc, "ImageUnpack", mid, &res)
	})
}

func (m *MetaProxy) ImageDelete(mid, rid string) error {
	return m.Call(mid, func(cli client.Client, svc map[string]string) error {
		return cli.Call(m.Context, m.svc, "ImageDelete", &DeleteReq{
			MetaId:  mid,
			ImageId: rid,
		}, nil)
	})
}

func (m *MetaProxy) ImageList(mid string, f func(string) error) error {
	return m.Call(mid, func(cli client.Client, svc map[string]string) error {
		res := []string{}
		err := cli.Call(m.Context, m.svc, "ImageList", mid, &res)
		if err != nil {
			return err
		}

		for _, v := range res {
			if err := f(v); err != nil {
				return err
			}
		}

		return nil
	})
}

func (m *MetaProxy) ImageAvailable(ctx context.Context, f func(string, string, []string) error) error {
	return m.Broadcast(func(cli client.Client) error {
		res := []Image{}

		err := cli.Call(m.Context, m.svc, "ImageAvailable", nil, &res)
		if err != nil {
			return err
		}

		for _, v := range res {
			if err := f(v.Id, v.Name, v.Refs); err != nil {
				return err
			}
		}

		return nil
	})
}
