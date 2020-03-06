package main

import (
	"context"
	"github.com/segmentio/ksuid"
	"github.com/xhebox/chrootd/api/container"
	"github.com/xhebox/chrootd/api/containerpool"
	"io"
	"log"
)

type ContainerServer struct {
	container.UnimplementedContainerServer
}

func NewContainerServer() *ContainerServer {
	return &ContainerServer{}
}

type PoolServer struct {
	containerpool.UnimplementedContainerPoolServer
	ContainerGroup map[string]string
}

func NewPoolServer() *PoolServer {
	return &PoolServer{
		ContainerGroup: make(map[string]string),
	}
}

func (co *ContainerServer) Handle(srv container.Container_HandleServer) error {
	log.Println("into handle")
	cnt := int32(0)
	for {
		log.Println("wait fot recv")
		data, err := srv.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		switch d := data.Payload.(type) {
		case *container.Packet_Id:
			cnt++
			log.Printf("open ContainerServer %v", d.Id)
		case *container.Packet_Data:
			cnt++
			log.Printf("get data : %v", d.Data)
		default:
			log.Printf("get unknown packet: %v", d)
		}
	}
	return srv.SendAndClose(&container.Reply{Seq: cnt, Code: 0, Message: "success", Type: "id"})
}

func (s *PoolServer) FindContainer(ctx context.Context, in *containerpool.Query) (*containerpool.Reply, error) {
	log.Println("find ContainerServer request")
	for key, value := range s.ContainerGroup {
		if value == in.Name {
			return &containerpool.Reply{Message: key, Code: 200}, nil
		}
	}

	if in.IsCreate {
		id := ksuid.New().String()
		s.ContainerGroup[id] = in.Name
		// TODO: create ContainerServer ...
		return &containerpool.Reply{Message: id, Code: 200}, nil
	}

	return &containerpool.Reply{Message: "not found", Code: 400}, nil
}

func (s *PoolServer) SetContainer(ctx context.Context, in *containerpool.SetRequest) (*containerpool.Reply, error) {
	log.Println("set ContainerServer request")
	switch body := in.Body.(type) {
	case *containerpool.SetRequest_Delete:
		//TODO: delete ContainerServer ...
		if _, ok := s.ContainerGroup[body.Delete.Id]; !ok {
			return &containerpool.Reply{Message: "ContainerServer does not exist", Code: 400}, nil
		}

		delete(s.ContainerGroup, body.Delete.Id)
		return &containerpool.Reply{Message: "delete ContainerServer successfully", Code: 200}, nil
	default:
		return &containerpool.Reply{Message: "nothing", Code: 200}, nil
	}
}
