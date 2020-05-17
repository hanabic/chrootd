package utils

import (
	"fmt"
	"net"
	"strconv"
)

type Addr struct {
	network string
	addr    string
	port    int
}

func NewAddr(network, addr string, port int) *Addr {
	return &Addr{
		network: network,
		addr:    addr,
		port:    port,
	}
}

func NewAddrFree() *Addr {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil
	}
	defer listener.Close()

	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return nil
	}

	return NewAddr("tcp", addr.IP.String(), addr.Port)
}

func NewAddrNet(addr net.Addr) *Addr {
	host, portstr, _ := net.SplitHostPort(addr.String())
	port, _ := strconv.Atoi(portstr)
	return NewAddr(addr.Network(), host, port)
}

func NewAddrString(network, addr string) *Addr {
	host, portstr, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portstr)
	return NewAddr(network, host, port)
}

func (c *Addr) SetNetwork(network string) {
	c.network = network
}

func (c *Addr) Network() string {
	return c.network
}

func (c *Addr) Addr() string {
	return c.addr
}

func (c *Addr) Port() int {
	return c.port
}

func (c *Addr) String() string {
	switch c.network {
	case "http", "https":
		if c.port <= 0 {
			return fmt.Sprintf("%s://%s/", c.network, c.addr)
		} else {
			return fmt.Sprintf("%s://%s:%d/", c.network, c.addr, c.port)
		}
	default:
		return fmt.Sprintf("%s:%d", c.addr, c.port)
	}
}
