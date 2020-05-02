package client

/*
import (
	"fmt"

	valkeyrie "github.com/abronan/valkeyrie/store"
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/store"
)

var (
	RegistryFlags = []cli.Flag{
		&cli.StringSliceFlag{
			Name:  "registry",
			Usage: "libkv store `addr`, used for container or service registry",
		},
		&cli.StringFlag{
			Name:  "registry_backend",
			Usage: "specify the `backend` used by registry, boltdb/zk/consul/etcdv3",
		},
		&cli.StringFlag{
			Name:  "registry_bucket",
			Usage: "if registry is backened by bolt, a `bucket` name is needed",
			Value: "chrootd",
		},
	}
)

func NewRegistryFromCli(c *cli.Context) (Registry, error) {
	backend := c.String("registry_backend")
	if len(backend) == 0 {
		return nil, fmt.Errorf("no backend set")
	}

	return store.NewValkeyrie(
		valkeyrie.Backend(backend),
		c.StringSlice("registry"),
		&valkeyrie.Config{Bucket: c.String("registry_bucket")},
	)
}
*/
