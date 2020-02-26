package server

import (
	"context"
	"errors"
	"github.com/segmentio/ksuid"
	. "github.com/xhebox/chrootd/api/containerpool/protobuf"
)

type PoolServer struct {
	UnimplementedContainerPoolServer
	ContainerGroup map[string]string
}

func NewPoolServer() *PoolServer {
	return &PoolServer{
		ContainerGroup: make(map[string]string),
	}
}

func (s *PoolServer) FindContainer(ctx context.Context, in *Query) (*Reply, error) {
	for key, value := range s.ContainerGroup {
		if value == in.Name {
			return &Reply{Message: key, Code: 200}, nil
		}
	}

	if in.IsCreate {
		id := ksuid.New().String()
		s.ContainerGroup[id] = in.Name
		//TODO: create container ...
		return &Reply{Message: id, Code: 200}, nil
	}
	return &Reply{Message: "not found", Code: 400}, nil
}

func (s *PoolServer) SetContainer(ctx context.Context, in *SetRequest) (*Reply, error) {
	switch in.State {
	case StateType_Delete:
		//TODO: delete container ...
		if _, ok := s.ContainerGroup[in.GetDelete().Id]; !ok {
			return &Reply{}, errors.New("container " + in.GetDelete().Id + " doesn't exits")
		}

		delete(s.ContainerGroup, in.GetDelete().Id)
	}
	return &Reply{Message: "delete container " + in.GetDelete().Id + " successfully", Code: 200}, nil
}
