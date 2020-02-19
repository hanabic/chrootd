package main

import (
	"context"
	"io"
	"log"
	"time"

	pb "github.com/xhebox/chrootd/api/api"
	"google.golang.org/grpc"
)

const (
	address     = "127.0.0.1:8999"
	defaultName = "world"
)

func runCommand(client pb.ApiClient, cmd string) *pb.Reply {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := client.RunCommand(ctx, &pb.Command{Cmd: cmd})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	} else {
		log.Printf("Route summary: %v", r)
	}
	return r
}

func runListConfig(client pb.ApiClient, id pb.Id) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.ListConfig(ctx, &id)
	if err != nil {
		log.Fatalf("%v.ListConfig(_) = _, %v", client, err)
	}
	var set []*pb.Config = nil

	for {
		config, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.ListConfig(_) = _, %v", client, err)
		}
		set = append(set, config)
	}
	log.Println("config:", set)
}

func runRecordConfig(client pb.ApiClient, config pb.Config) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.RecordConfig(ctx)
	if err != nil {
		log.Fatalf("%v.RecordRoute(_) = _, %v", client, err)
	}

	if err := stream.Send(&config); err != nil {
		log.Fatalf("%v.Send(%v) = %v", stream, config, err)
	}

	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
	}
	log.Printf("Route summary: %v", reply)
}

var conf = pb.Config{
	Id:    "a02",
	Path:  "/home",
	Level: "low",
}

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewApiClient(conn)

	runCommand(client, "cd")

	runListConfig(client, pb.Id{Id: "a01"})

	runRecordConfig(client, conf)

}
