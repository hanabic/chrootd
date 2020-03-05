package server

import (
	. "github.com/xhebox/chrootd/api/container/protobuf"
	"io"
	"log"
)

type container struct {
	UnimplementedContainerServer
}

func NewContainerServer() *container {
	return &container{}
}

func (co *container) Handle(srv Container_HandleServer) error {
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
		case *Packet_Id:
			cnt++
			log.Printf("open container %v", d.Id)
		case *Packet_Data:
			cnt++
			log.Printf("get data : %v", d.Data)
		default:
			log.Printf("get unknown packet: %v", d)
		}
	}
	return srv.SendAndClose(&Reply{Seq: cnt, Code: 0, Message: "success", Type: "id"})
}
