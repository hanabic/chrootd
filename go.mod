module github.com/xhebox/chrootd

go 1.13

replace github.com/xhebox/chrootd => ./

require (
	github.com/go-ini/ini v1.55.0
	github.com/gobwas/glob v0.2.3
	github.com/golang/protobuf v1.4.0-rc.4
	github.com/opencontainers/runc v1.0.0-rc9.0.20200316180000-939cd0b734a0
	github.com/pkg/errors v0.8.1
	github.com/rs/zerolog v1.18.0
	github.com/segmentio/ksuid v1.0.2
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/urfave/cli/v2 v2.2.0
	golang.org/x/sys v0.0.0-20200217220822-9197077df867
	google.golang.org/grpc v1.27.1
	google.golang.org/protobuf v1.20.1
	gopkg.in/ini.v1 v1.55.0 // indirect
)
