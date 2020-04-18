package utils

import (
	"fmt"
	"strings"
)

type Addr struct {
	network string
	addr    string
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
