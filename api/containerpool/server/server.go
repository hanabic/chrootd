package server

import (
	"context"
	"log"

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
	log.Println("find container request")
	for key, value := range s.ContainerGroup {
		if value == in.Name {
			return &Reply{Message: key, Code: 200}, nil
		}
	}

	if in.IsCreate {
		id := ksuid.New().String()
		s.ContainerGroup[id] = in.Name
		// TODO: create container ...
		return &Reply{Message: id, Code: 200}, nil
	}

	return &Reply{Message: "not found", Code: 400}, nil
}

func (s *PoolServer) SetContainer(ctx context.Context, in *SetRequest) (*Reply, error) {
	log.Println("set container request")
	switch body := in.Body.(type) {
	case *SetRequest_Delete:
		//TODO: delete container ...
		if _, ok := s.ContainerGroup[body.Delete.Id]; !ok {
			return &Reply{Message: "container does not exist", Code: 400}, nil
		}

		delete(s.ContainerGroup, body.Delete.Id)
		return &Reply{Message: "delete container successfully", Code: 200}, nil
	default:
		return &Reply{Message: "nothing", Code: 200}, nil
	}
}
