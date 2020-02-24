module github.com/xhebox/chrootd

go 1.13

replace github.com/xhebox/chrootd/commands => ./commands

require (
	github.com/go-ini/ini v1.52.0
	github.com/golang/protobuf v1.3.2
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/sevlyar/go-daemon v0.1.5
	golang.org/x/sys v0.0.0-20200217220822-9197077df867 // indirect
	google.golang.org/grpc v1.27.1
	gopkg.in/ini.v1 v1.52.0 // indirect
)
