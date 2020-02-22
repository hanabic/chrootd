package common

import (
	"flag"
	"fmt"
	"net"

	"google.golang.org/grpc"
)

type GrpcConfig struct {
	Url         string
	NetWorkType string
}

func (conf *GrpcConfig) SetFlag(fs *flag.FlagSet) {
	fs.StringVar(&conf.Url, "url", "127.0.0.1:9090", "host of grpc")
	fs.StringVar(&conf.NetWorkType, "type", "tcp", "type of grpc")
}

func (conf *GrpcConfig) Connect(opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.Dial(conf.Url, opts...)
}

func (conf *GrpcConfig) RunServer(opts ...grpc.ServerOption) error {
	lis, err := net.Listen(conf.NetWorkType, conf.Url)
	if err != nil {
		return fmt.Errorf("grpc: failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(opts...)
	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("grpc: failed to serve: %v", err)
	}

	return nil
}
