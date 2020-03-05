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
	cnt := 0
	for {
		log.Println("wait fot recv")
		data, err := srv.Recv()
		if err == io.EOF {
			return srv.SendMsg(&Reply{Message: "success", Code: 200})
		}
		if err != nil {
			return err
		}
		switch d := data.Payload.(type) {
		case *Packet_Id:
			cnt++
			log.Printf("open start %v", d.Id)

		case *Packet_Data:
			cnt++
			log.Printf("get data : %v", d.Data)

		case *Packet_End:
			cnt++
			log.Printf("end")
			if err := srv.SendAndClose(&Reply{Seq: int32(cnt), Code: 0, Message: "success", Type: "end"}); err == nil {
				return err
			}
			return nil
		}
	}
}
