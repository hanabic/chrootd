module github.com/xhebox/chrootd

go 1.13

replace github.com/xhebox/chrootd => ./

require (
	github.com/asaskevich/govalidator v0.0.0-20200108200545-475eaeb16496
	github.com/boltdb/bolt v1.3.1 // indirect
	github.com/containerd/console v1.0.0 // indirect
	github.com/cornelk/hashmap v1.0.1 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/go-units v0.4.0
	github.com/docker/libkv v0.2.1
	github.com/etcd-io/etcd v3.3.20+incompatible // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/golang/protobuf v1.4.0-rc.4
	github.com/imdario/mergo v0.3.9
	github.com/micro/go-micro v1.18.0 // indirect
	github.com/micro/go-micro/v2 v2.4.0
	github.com/opencontainers/runc v1.0.0-rc9.0.20200316180000-939cd0b734a0
	github.com/pelletier/go-toml v1.7.0
	github.com/pkg/errors v0.9.1
	github.com/pkg/xattr v0.4.1
	github.com/rs/zerolog v1.18.0
	github.com/ryanuber/go-glob v1.0.0
	github.com/segmentio/ksuid v1.0.2
	github.com/smallnest/libkv-etcdv3-store v1.1.1
	github.com/smallnest/rpcx v0.0.0-20200414114925-bff251b691b9
	github.com/spf13/cobra v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/urfave/cli/v2 v2.2.0
	github.com/xhebox/libkv v0.2.1 // indirect
	github.com/xhebox/libkv-bolt v0.0.0-20200418115446-d20436404098
	golang.org/x/net v0.0.0-20200222125558-5a598a2470a0
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sys v0.0.0-20200217220822-9197077df867
	google.golang.org/grpc v1.26.0
)
