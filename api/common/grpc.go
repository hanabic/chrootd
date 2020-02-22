package common

import (
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
)

type GrpcConfig struct {
	Url         string
	NetWorkType string
}

func (conf *GrpcConfig) Parse(fs *flag.FlagSet) {
	fs.StringVar(&conf.Url, "url", "127.0.0.1:9090", "host of grpc")
	fs.StringVar(&conf.NetWorkType, "type", "tcp", "type of grpc")
}

func (conf *GrpcConfig) Connect(opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.Dial(conf.Url, opts...)
}

func (conf *GrpcConfig) RunServer(opts ...grpc.ServerOption) *grpc.Server {
	lis, err := net.Listen(conf.NetWorkType, conf.Url)
	if err != nil {
		log.Fatalf("grpc: failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer(opts...)
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("grpc: failed to serve: %v", err)
	}
	fmt.Println("grpc: server start")
	return grpcServer
}
