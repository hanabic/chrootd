package main

import (
	pb "github.com/xhebox/chrootd/api/container"
	"io"
	"sync"
)

type container struct {
	id     string
	name   string
	statue string
}

type containerService struct {
	pb.UnimplementedContainerServer
	containerMap *sync.Map
}

func newContainer(id string, name string) *container {
	return &container{id: id, name: name, statue: "close"}
}

func newContainerService() *containerService {
	return &containerService{
		containerMap: new(sync.Map),
	}
}

func (service *containerService) Handle(srv pb.Container_HandleServer) error {
	cnt := int32(0)
	var currentId string
	for {
		data, err := srv.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		cnt++
		switch d := data.Payload.(type) {
		case *pb.Packet_Id:
			currentId = d.Id
		default:
			if _, ok := (*service.containerMap).Load(currentId); ok {
			} else {
				return srv.SendAndClose(&pb.Reply{Seq: cnt, Code: 400, Message: "container not exits"})
			}
		}
	}
	return srv.SendAndClose(&pb.Reply{Seq: cnt, Code: 200, Message: "success", Type: "id"})
}
