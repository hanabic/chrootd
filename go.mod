module github.com/xhebox/chrootd

go 1.13

replace github.com/xhebox/chrootd => ./

require (
	github.com/checkpoint-restore/go-criu v0.0.0-20191125063657-fcdcd07065c5 // indirect
	github.com/containerd/console v1.0.0 // indirect
	github.com/docker/go-units v0.4.0
	github.com/godbus/dbus v0.0.0-20190422162347-ade71ed3457e // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.3.5 // indirect
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/hashicorp/consul/api v1.4.0
	github.com/hashicorp/consul/sdk v0.4.0
	github.com/imdario/mergo v0.3.9
	github.com/klauspost/compress v1.10.3 // indirect
	github.com/klauspost/pgzip v1.2.3 // indirect
	github.com/mrunalp/fileutils v0.0.0-20171103030105-7d4729fb3618 // indirect
	github.com/openSUSE/umoci v0.4.5
	github.com/opencontainers/image-spec v1.0.2-0.20190823105129-775207bd45b6
	github.com/opencontainers/runc v1.0.0-rc9.0.20200514005706-3f1e88699199
	github.com/opencontainers/runtime-spec v1.0.2
	github.com/opencontainers/selinux v1.5.1 // indirect
	github.com/osamingo/jsonrpc v0.0.0-20200219081550-352acaa9f2b2
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pelletier/go-toml v1.7.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/procfs v0.0.5 // indirect
	github.com/rs/zerolog v1.18.0
	github.com/seccomp/libseccomp-golang v0.9.1 // indirect
	github.com/segmentio/ksuid v1.0.2
	github.com/sirupsen/logrus v1.6.0 // indirect
	github.com/smallnest/rpcx v0.0.0-20200512151426-9e5976a9d1d6
	github.com/stretchr/testify v1.5.1 // indirect
	github.com/syndtr/gocapability v0.0.0-20180916011248-d98352740cb2 // indirect
	github.com/tidwall/gjson v1.6.0
	github.com/urfave/cli/v2 v2.2.0
	github.com/vishvananda/netlink v1.1.0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190809123943-df4f5c81cb3b // indirect
	github.com/ybbus/jsonrpc v1.1.2-0.20200212073916-a94e6ce5643c
	go.etcd.io/bbolt v1.3.4
	golang.org/x/crypto v0.0.0-20200510223506-06a226fb4e37
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sys v0.0.0-20200327173247-9dae0f8f5775
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	google.golang.org/grpc v1.24.0 // indirect
	gopkg.in/oauth2.v3 v3.12.0
)
