package main

import (
	"context"
	"encoding/json"
	"fmt"
	pb "github.com/xhebox/chrootd/api/api"
	"google.golang.org/grpc"
	"io"
	"io/ioutil"
	"log"
	"net"
)

type Server struct {
	pb.UnimplementedApiServer
	config []*pb.Config
}

type ServerOptions struct {
	address    string
	netType    string
	configPath string
}

var exampleConfig = []byte(`[{
	"id": "a01" ,
	"path":"/use/bin",
	"level":"high"
}]`)

func (s *Server) LoadConfig(filePath string) {
	var data []byte
	if filePath != "" {
		var err error
		data, err = ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Failed to load default config: %v", err)
		}
	} else {
		data = exampleConfig
	}

	if err := json.Unmarshal(data, &s.config); err != nil {
		log.Fatalf("Failed to load default config: %v", err)
	}

}

func (s *Server) RunCommand(ctx context.Context, cmd *pb.Command) (*pb.Reply, error) {

	log.Println(cmd)

	/*todo: check and run command*/

	return &pb.Reply{Message: "success", Code: 200}, nil
}

func (s *Server) ListConfig(id *pb.Id, stream pb.Api_ListConfigServer) error {
	for _, config := range s.config {
		fmt.Println(config)
		if config.Id == id.Id {

			if err := stream.Send(config); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Server) RecordConfig(stream pb.Api_RecordConfigServer) error {
	for {
		data, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.Reply{Message: "success", Code: 200})
		}
		if err != nil {
			return err
		}
		log.Println(data)
	}
}

func runServer(opt ServerOptions) *grpc.Server {
	lis, err := net.Listen(opt.netType, opt.address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	srv := &Server{}
	srv.LoadConfig(opt.configPath)
	pb.RegisterApiServer(grpcServer, srv)
	grpcServer.Serve(lis)
	return grpcServer
}

func main() {
	opt := ServerOptions{
		address:    "127.0.0.1:8999",
		netType:    "tcp",
		configPath: "",
	}
	runServer(opt)
}
