package client

import (
	"context"
	. "github.com/xhebox/chrootd/api/container/protobuf"
	"log"
	"time"
)

func StartContainer(client ContainerClient, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.Handle(ctx)
	if err != nil {
		log.Fatalf("%v.RecordRoute(_) = _, %v", client, err)
	}

	log.Printf("start container %v\n", id)

	// send id packet
	if err := stream.Send(&Packet{Payload: &Packet_Id{Id: "ddddd"}}); err != nil {
		log.Println("error in send id packet")
		return err
	}
	//send end packet
	if err := stream.Send(&Packet{Payload: &Packet_End{End: "end"}}); err != nil {
		log.Println("error in send end packet")
		return err
	}

	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Println("error in close stream")
		return err
	}
	log.Printf("Reply summary: %v", reply)
	return nil
}
