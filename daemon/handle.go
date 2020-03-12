package main

import (
	"github.com/xhebox/chrootd/api/container"
	"io"
	"log"
)

type SingleContainer struct {
	id     string
	name   string
	statue string
}

type ContainerServer struct {
	Container.UnimplementedContainerServer
	group *map[string]*SingleContainer
}

func NewContainerServer() *ContainerServer {
	return &ContainerServer{}
}

func (s *SingleContainer) enqueue() error {
	log.Println("hddd")
	return nil
}

func (co *ContainerServer) Handle(srv Container.Container_HandleServer) error {
	log.Println("into handle")
	cnt := int32(0)
	//var currentId string
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
		case *Container.Packet_Id:
			//currentId = d.Id
			log.Printf("get id : %v", d.Id)
		case *Container.Packet_Start:
			log.Println("get start packet")
		case *Container.Packet_Stop:
			log.Println("get stop packet")
		default:
			log.Printf("get unknown packet: %v", d)
		}
	}
	//og.Println("finish")
	return srv.SendAndClose(&Container.Reply{Seq: cnt, Code: 0, Message: "success", Type: "id"})
}
