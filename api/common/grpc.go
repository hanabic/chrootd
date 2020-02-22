package common

import (
	"flag"
	"net"
	"time"
)

type ConnConfig struct {
	Url         string
	NetWorkType string
	Timeout     time.Duration
}

func (conf *ConnConfig) SetFlag(fs *flag.FlagSet) {
	fs.StringVar(&conf.Url, "url", "127.0.0.1:9090", "host of grpc")
	fs.StringVar(&conf.NetWorkType, "type", "tcp", "type of conn")
	fs.DurationVar(&conf.Timeout, "timeout", 30*time.Second, "dial timeout")
}

func (conf *ConnConfig) Dial() (net.Conn, error) {
	return net.DialTimeout(conf.NetWorkType, conf.Url, conf.Timeout)
}

func (conf *ConnConfig) Listen() (net.Listener, error) {
	return net.Listen(conf.NetWorkType, conf.Url)
}
