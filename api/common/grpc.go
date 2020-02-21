package common

import "flag"

type GrpcConfig struct {
	Url         string
	NetWorkType string
}

func (conf *GrpcConfig) Parse(fs *flag.FlagSet) {
	fs.StringVar(&conf.Url, "url", "127.0.0.1:9090", "host of grpc")
	fs.StringVar(&conf.NetWorkType, "type", "tcp", "type of grpc")
}

