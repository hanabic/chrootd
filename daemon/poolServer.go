package main

import (
	"context"
	"errors"

	"github.com/gobwas/glob"
	"github.com/xhebox/chrootd/api"
	bolt "go.etcd.io/bbolt"
)

type poolServer struct {
	api.UnimplementedContainerPoolServer
	user  *User
	cntrs *cntrPool
}

func newPoolServer(u *User, p *cntrPool) *poolServer {
	return &poolServer{user: u, cntrs: p}
}

func (s *poolServer) Create(ctx context.Context, req *api.CreateReq) (*api.CreateRes, error) {
	s.user.Logger.Info().Msgf("request to create container[%p]\n%v", req, string(req.Config))

	id, err := s.cntrs.Add(req.Config)
	if err != nil {
		// TODO: log error, return detailed error
		s.user.Logger.Error().Msgf("fail to create: %v", err)
		return &api.CreateRes{Id: nil, Reason: "fail to create"}, nil
	}

	return &api.CreateRes{Id: id}, nil
}

func (s *poolServer) Update(ctx context.Context, req *api.UpdateReq) (*api.UpdateRes, error) {
	s.user.Logger.Info().Msgf("request to update container[%p]", req)

	res, err := s.cntrs.UpdateMeta(req.Id, req.Config)
	if err != nil {
		// TODO: log error, return detailed error
		return &api.UpdateRes{Config: nil, Reason: "failed"}, nil
	}

	return &api.UpdateRes{Config: res}, nil
}

func (s *poolServer) Delete(ctx context.Context, req *api.DeleteReq) (*api.DeleteRes, error) {
	s.user.Logger.Info().Msgf("request to delete container[%p]", req)

	if err := s.cntrs.Del(req.Id); err != nil {
		// TODO: log error, return detailed error
		return &api.DeleteRes{Reason: "failed"}, nil
	}

	return &api.DeleteRes{}, nil
}

func (s *poolServer) List(req *api.ListReq, srv api.ContainerPool_ListServer) error {
	s.user.Logger.Info().Msgf("request to list containers[%p]", req)

	var err error

	var nameGlob, labelGlob, idGlob glob.Glob

	s.cntrs.ForEach(func(id []byte, b *bolt.Bucket) error {
		metadata := b.Get([]byte("metadata"))
		m, err := api.NewMetaFromBytes(metadata)
		if err != nil {
			return err
		}

		r := true

		for _, filter := range req.Filters {
			switch filter.Key {
			case "name":
				if nameGlob == nil {
					nameGlob, err = glob.Compile(filter.Val)
					if err != nil {
						return err
					}
				}

				if !nameGlob.Match(m.Name) {
					r = false
					break
				}
			case "label":
				if labelGlob == nil {
					labelGlob, err = glob.Compile(filter.Val)
					if err != nil {
						return err
					}
				}

				isBreak := false

				for _, v := range m.Config.Labels {
					if !labelGlob.Match(v) {
						r = false
						isBreak = true
						break
					}
				}
				if isBreak {
					break
				}
			case "id":
				if idGlob == nil {
					idGlob, err = glob.Compile(filter.Val)
					if err != nil {
						return err
					}
				}

				if !idGlob.Match(string(m.Id)) {
					r = false
					break
				}
			case "hostname":
				if nameGlob == nil {
					nameGlob, err = glob.Compile(filter.Val)
					if err != nil {
						return err
					}
				}

				if !nameGlob.Match(string(m.Config.Hostname)) {
					r = false
					break
				}
			default:
				return errors.New("unknown, stop")
			}
		}

		if r {
			cfg, err := m.ToBytes()
			if err != nil {
				return err
			}

			if err := srv.Send(&api.ListRes2{
				Id:     id,
				Config: cfg,
			}); err != nil {
				return err
			}
		}

		return nil
	})

	return err
}
