package main

import (
	"context"

	"github.com/gobwas/glob"
	"github.com/segmentio/ksuid"
	"github.com/xhebox/chrootd/api"
)

type poolServer struct {
	api.UnimplementedContainerPoolServer
	user  *User
	cntrs *cntrPool
}

func newPoolServer(u *User, p *cntrPool) *poolServer {
	return &poolServer{user: u, cntrs: p}
}

func (s *poolServer) Close() {
	s.cntrs.Range(func(key ksuid.KSUID, t *container) bool {
		t.Destroy()
		return true
	})
}

func (s *poolServer) Create(ctx context.Context, req *api.CreateReq) (*api.CreateRes, error) {
	s.user.Logger.Info().Msgf("request to create container[%p] %v", req, req)

	realCntr := newCntr(req.Container)

	id := s.cntrs.Add(realCntr)
	if id == ksuid.Nil {
		return &api.CreateRes{Id: nil, Reason: "can not alloc uid"}, nil
	}

	return &api.CreateRes{Id: id.Bytes()}, nil
}

func (s *poolServer) Update(ctx context.Context, req *api.UpdateReq) (*api.UpdateRes, error) {
	s.user.Logger.Info().Msgf("request to update container[%p]", req)

	id, err := ksuid.FromBytes(req.Id)
	if err != nil {
		return &api.UpdateRes{Container: nil, Reason: "invalid id"}, nil
	}

	cntr := s.cntrs.Get(id)
	if cntr == nil {
		return &api.UpdateRes{Container: nil, Reason: "no container of such id"}, nil
	}

	cntr.UpdateMeta(req.Container)

	return &api.UpdateRes{Container: cntr.Container}, nil
}

func (s *poolServer) Delete(ctx context.Context, req *api.DeleteReq) (*api.DeleteRes, error) {
	s.user.Logger.Info().Msgf("request to delete container[%p]", req)

	id, err := ksuid.FromBytes(req.Id)
	if err != nil {
		return &api.DeleteRes{Reason: "invalid id"}, nil
	}

	cntr := s.cntrs.Get(id)
	if cntr == nil {
		return &api.DeleteRes{Reason: "no container of such id"}, nil
	}

	if cntr.IsBusy() {
		return &api.DeleteRes{Reason: "container busy"}, nil
	}

	s.cntrs.Del(id)

	return &api.DeleteRes{}, nil
}

func (s *poolServer) List(req *api.ListReq, srv api.ContainerPool_ListServer) error {
	s.user.Logger.Info().Msgf("request to list containers[%p]", req)

	var err error

	var nameGlob glob.Glob

	s.cntrs.Range(func(key ksuid.KSUID, cntr *container) bool {
		r := true

		for _, filter := range req.Filters {
			switch filter.Key {
			// TODO: more filters
			case "name":
				if nameGlob == nil {
					nameGlob, err = glob.Compile(filter.Val)
					if err != nil {
						return false
					}
				}

				if !nameGlob.Match(cntr.Name) {
					r = false
					break
				}
			default:
				r = false
				return false
			}
		}

		if r {
			if e := srv.Send(cntr.Container); e != nil {
				err = e
				return false
			}
		}

		return true
	})

	return err
}
