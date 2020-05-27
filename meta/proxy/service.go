package proxy

import (
	"context"

	"github.com/hashicorp/consul/api"
	"github.com/segmentio/ksuid"
	"github.com/pkg/errors"
	mtyp "github.com/xhebox/chrootd/meta"
	"github.com/xhebox/chrootd/utils"
)

type MetaService struct {
	id          string
	addr        *utils.Addr
	reg         *api.AgentServiceRegistration
	cli         *api.Client
	mgr         mtyp.Manager
	QueryLimits int
}

func NewMetaService(mgr mtyp.Manager, cli *api.Client, svcname string, rpcAddr *utils.Addr) (*MetaService, error) {
	svc := &MetaService{
		cli:         cli,
		mgr:         mgr,
		addr:        rpcAddr,
		QueryLimits: 64,
	}

	if cli != nil {
		id, err := mgr.ID()
		if err != nil {
			return nil, err
		}
		svc.id = id

		svc.reg = &api.AgentServiceRegistration{
			ID:      ksuid.New().String(),
			Name:    svcname,
			Address: svc.addr.Addr(),
			Port:    svc.addr.Port(),
			Tags:    []string{svc.id},
		}

		err = cli.Agent().ServiceRegister(svc.reg)
		if err != nil {
			return nil, err
		}
	}

	return svc, nil
}

func (s *MetaService) ID(ctx context.Context, req *struct{}, res *string) error {
	*res = s.id
	return nil
}

func (s *MetaService) Create(ctx context.Context, meta *mtyp.Metainfo, res *string) error {
	id, err := s.mgr.Create(meta)
	if err == nil {
		*res = id
	}
	return err
}

func (s *MetaService) Get(ctx context.Context, id string, res *mtyp.Metainfo) error {
	meta, err := s.mgr.Get(id)
	if err == nil {
		*res = *meta
	}
	return err
}

func (s *MetaService) Update(ctx context.Context, req *mtyp.Metainfo, res *struct{}) error {
	return s.mgr.Update(req)
}

func (s *MetaService) Delete(ctx context.Context, cid string, res *struct{}) error {
	return s.mgr.Delete(cid)
}

func (s *MetaService) Query(ctx context.Context, query string, res *[]*mtyp.Metainfo) error {
	cnt := 0
	return s.mgr.Query(query, func(meta *mtyp.Metainfo) error {
		*res = append(*res, meta)
		cnt++
		if cnt > s.QueryLimits {
			return errors.New("exceed the query limits")
		}
		return nil
	})
}

func (s *MetaService) ImageUnpack(ctx context.Context, req string, res *string) error {
	var err error
	*res, err = s.mgr.ImageUnpack(ctx, req)
	return err
}

type DeleteReq struct {
	MetaId  string
	ImageId string
}

func (s *MetaService) ImageDelete(ctx context.Context, req *DeleteReq, res *struct{}) error {
	return s.mgr.ImageDelete(req.MetaId, req.ImageId)
}

func (s *MetaService) ImageList(ctx context.Context, cid string, res *[]string) error {
	cnt := 0
	return s.mgr.ImageList(cid, func(id string) error {
		*res = append(*res, id)
		cnt++
		if cnt > s.QueryLimits {
			return errors.New("exceed the query limits")
		}
		return nil
	})
}

type Image struct {
	Id   string
	Name string
	Refs []string
}

func (s *MetaService) ImageAvailable(ctx context.Context, req struct{}, res *[]Image) error {
	cnt := 0
	return s.mgr.ImageAvailable(ctx, func(id string, name string, refs []string) error {
		*res = append(*res, Image{Id: id, Name: name, Refs: refs})
		cnt++
		if cnt > s.QueryLimits {
			return errors.New("exceed the query limits")
		}
		return nil
	})
}
