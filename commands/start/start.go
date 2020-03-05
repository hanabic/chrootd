package start

import (
	"context"
	"flag"
	"log"
	"net"

	. "github.com/xhebox/chrootd/api/container/client"
	pb "github.com/xhebox/chrootd/api/container/protobuf"
	. "github.com/xhebox/chrootd/commands"
	. "github.com/xhebox/chrootd/common"
	"google.golang.org/grpc"
)

var Container = Command{
	Name: "start",
	Desc: "start start by id",
	Hanlder: func(args []string) error {
		fs := flag.NewFlagSet("start", flag.ContinueOnError)
		connConf := ConnConfig{}
		connConf.SetFlag(fs)
		connConf.LoadEnv()

		if err := fs.Parse(args); err != nil {
			return err
		}

		log.Printf("connecting to grpc server %s via %s\n", connConf.ContainerAddr, connConf.NetWorkType)

		conn, err := grpc.Dial("start", grpc.WithInsecure(), grpc.WithContextDialer(func(ctx context.Context, target string) (net.Conn, error) {
			return connConf.ContainerDial()
		}))

		defer conn.Close()

		client := pb.NewContainerClient(conn)
		if err := StartContainer(client, args[0]); err != nil {
			return err
		}
		return err
	},
}
