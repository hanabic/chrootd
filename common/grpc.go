package common

import (
	"flag"
	"net"
	"os"
	"strconv"
	"time"
)

type ConnConfig struct {
	PoolAddr      string
	ContainerAddr string
	NetWorkType   string
	Timeout       time.Duration
}

func (conf *ConnConfig) SetFlag(fs *flag.FlagSet) {
	fs.StringVar(&conf.PoolAddr, "pooladdr", "127.0.0.1:9090", "pool server connection addr")
	fs.StringVar(&conf.PoolAddr, "containeraddr", "127.0.0.1:8090", "start connection addr")
	fs.StringVar(&conf.NetWorkType, "conntype", "tcp", "connection type")
	fs.DurationVar(&conf.Timeout, "conntimeout", 30*time.Second, "connection dial timeout")
}

func (conf *ConnConfig) PoolDial() (net.Conn, error) {
	return net.DialTimeout(conf.NetWorkType, conf.PoolAddr, conf.Timeout)
}

func (conf *ConnConfig) ContainerDial() (net.Conn, error) {
	return net.DialTimeout(conf.NetWorkType, conf.ContainerAddr, conf.Timeout)
}

func (conf *ConnConfig) PoolListen() (net.Listener, error) {
	return net.Listen(conf.NetWorkType, conf.PoolAddr)
}

func (conf *ConnConfig) ContainerListen() (net.Listener, error) {
	return net.Listen(conf.NetWorkType, conf.ContainerAddr)
}

func (conf *ConnConfig) LoadEnv() {
	if value := os.Getenv("CHROOTD_CONNADDR"); value != "" {
		conf.PoolAddr = value
	}

	if value := os.Getenv("CHROOTD_CONTAINERADDR"); value != "" {
		conf.ContainerAddr = value
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
