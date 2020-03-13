package main

import (
	"context"
	"github.com/segmentio/ksuid"
	"log"
	"sync"

	pb "github.com/xhebox/chrootd/api/containerpool"
)

type pool struct {
	statue string
}

func newPool() *pool {
	return &pool{}
}

type poolService struct {
	pb.UnimplementedContainerPoolServer
	ContainerGroup sync.Map
}

func newPoolService() *poolService {
	return &poolService{}
}

func (s *poolService) FindContainer(ctx context.Context, in *pb.Query) (*pb.Reply, error) {
	log.Println("find containerService request")
	result := &pb.Reply{
		Message: "not found",
		Code:    400,
	}

	s.ContainerGroup.Range(func(key, value interface{}) bool {
		if value == in.Name {
			result.Message = key.(string)
			result.Code = value.(int32)
		}
		return true
	})

	if in.IsCreate {
		id := ksuid.New().String()
		s.ContainerGroup.Store(id, in.Name)
		// TODO: create containerService ...
		return &pb.Reply{Message: id, Code: 200}, nil
	}

	return &pb.Reply{Message: "not found", Code: 400}, nil
}

func (s *poolService) SetContainer(ctx context.Context, in *pb.SetRequest) (*pb.Reply, error) {
	log.Println("set containerService request")
	switch body := in.Body.(type) {
	case *pb.SetRequest_Delete:
		//TODO: delete containerService ...
		if _, ok := s.ContainerGroup.Load(body.Delete.Id); !ok {
			return &pb.Reply{Message: "containerService does not exist", Code: 400}, nil
		}

		s.ContainerGroup.Delete(body.Delete.Id)
		return &pb.Reply{Message: "delete containerService successfully", Code: 200}, nil
	default:
		return &pb.Reply{Message: "nothing", Code: 200}, nil
	}
}
