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
		if err := fs.Parse(args); err != nil {
			return err
		}
		connConf.LoadEnv()

		log.Printf("connecting to grpc server %s via %s\n", connConf.ContainerAddr, connConf.NetWorkType)

		conn, err := grpc.Dial("start", grpc.WithInsecure(), grpc.WithContextDialer(func(ctx context.Context, target string) (net.Conn, error) {
			return connConf.ContainerDial()
		}))
		if err != nil {
			return err
		}

		defer conn.Close()

		client := pb.NewContainerClient(conn)
		return StartContainer(client, args[0])
	},
}
