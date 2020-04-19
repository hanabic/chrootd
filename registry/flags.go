package registry

import (
	"fmt"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/consul"
	"github.com/docker/libkv/store/zookeeper"
	etcdv3 "github.com/smallnest/libkv-etcdv3-store"
	"github.com/xhebox/libkv-bolt"
	"github.com/urfave/cli/v2"
)

func init() {
	boltdb.Register()
	etcdv3.Register()
	consul.Register()
	zookeeper.Register()
}

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

	store, err := libkv.NewStore(
		store.Backend(backend),
		c.StringSlice("registry"),
		&store.Config{Bucket: c.String("registry_bucket")},
	)
	if err != nil {
		return nil, err
	}

	return NewStoreRegistry(store), nil
}
