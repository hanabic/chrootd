module github.com/xhebox/chrootd

go 1.13

replace github.com/xhebox/chrootd => ./

require (
	github.com/containers/skopeo v0.2.0 // indirect
	github.com/docker/go-units v0.4.0
	github.com/hashicorp/consul/api v1.4.0
	github.com/hashicorp/consul/sdk v0.4.0
	github.com/imdario/mergo v0.3.9
	github.com/openSUSE/umoci v0.4.5
	github.com/opencontainers/image-spec v1.0.2-0.20190823105129-775207bd45b6
	github.com/opencontainers/runc v1.0.0-rc9.0.20200506213300-64416d34f30e
	github.com/opencontainers/runtime-spec v1.0.2
	github.com/osamingo/jsonrpc v0.0.0-20200219081550-352acaa9f2b2
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pelletier/go-toml v1.7.0
	github.com/pkg/errors v0.9.1
	github.com/pkg/xattr v0.4.1
	github.com/rs/zerolog v1.18.0
	github.com/segmentio/ksuid v1.0.2
	github.com/smallnest/rpcx v0.0.0-20200512151426-9e5976a9d1d6
	github.com/tidwall/gjson v1.6.0
	github.com/urfave/cli/v2 v2.2.0
	github.com/ybbus/jsonrpc v1.1.2-0.20200212073916-a94e6ce5643c
	go.etcd.io/bbolt v1.3.4
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sys v0.0.0-20200327173247-9dae0f8f5775
	gopkg.in/oauth2.v3 v3.12.0
)
