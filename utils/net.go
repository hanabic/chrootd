package utils

import (
	"fmt"
	"net"
	"strings"
)

type Addr struct {
	network string
	addr    string
}

func NewAddrFree() *Addr {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil
	}
	defer listener.Close()

	return &Addr{
		network: listener.Addr().Network(),
		addr:    listener.Addr().String(),
	}
}

func NewAddrFromString(s string) *Addr {
	r := strings.SplitN(s, "@", 2)
	if len(r) == 2 {
		return &Addr{
			network: r[0],
			addr:    r[1],
		}
	} else {
		return &Addr{
			network: "tcp",
			addr:    r[0],
		}
	}
}

func NewAddr(a, b string) *Addr {
	return &Addr{
		network: a,
		addr:    b,
	}
}

func (c *Addr) Network() string {
	return c.network
}

func (c *Addr) Addr() string {
	return c.addr
}

func (c *Addr) String() string {
	return fmt.Sprintf("%s@%s", c.network, c.addr)
}
