package common

import (
	"flag"
	"net"
	"os"
	"strconv"
	"time"
)

type ConnConfig struct {
	Addr        string
	NetWorkType string
	Timeout     time.Duration
}

func (conf *ConnConfig) SetFlag(fs *flag.FlagSet) {
	fs.StringVar(&conf.Addr, "connaddr", "127.0.0.1:9090", "connection addr")
	fs.StringVar(&conf.NetWorkType, "conntype", "tcp", "connection type")
	fs.DurationVar(&conf.Timeout, "conntimeout", 30*time.Second, "connection dial timeout")
}

func (conf *ConnConfig) Dial() (net.Conn, error) {
	return net.DialTimeout(conf.NetWorkType, conf.Addr, conf.Timeout)
}

func (conf *ConnConfig) Listen() (net.Listener, error) {
	return net.Listen(conf.NetWorkType, conf.Addr)
}

func (conf *ConnConfig) LoadEnv() {
	if value := os.Getenv("CHROOTD_CONNADDR"); value != "" {
		conf.Addr = value
	}
	if value := os.Getenv("CHROOTD_CONNTYPE"); value != "" {
		conf.NetWorkType = value
	}
	if value := os.Getenv("CHROOTD_CONNTIMEOUT"); value != "" {
		va, err := strconv.Atoi(value)
		if err == nil {
			conf.Timeout = time.Duration(va)
		}
	}
}
