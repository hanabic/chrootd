package start

import (
	"context"
	"flag"
	"github.com/xhebox/chrootd/api/container"
	"log"
	"net"
	"time"

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

		client := container.NewContainerClient(conn)
		return StartContainer(client, args[0])
	},
}

func StartContainer(client container.ContainerClient, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.Handle(ctx)
	if err != nil {
		return err
	}
	log.Printf("start container %v\n", id)

	if err := stream.Send(&container.Packet{Payload: &container.Packet_Id{Id: "ddddd"}}); err != nil {
		log.Println("error in send id packet")
		return err
	}

	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Println("error in close stream")
		return err
	}
	log.Printf("Reply summary: %v", reply)
	return nil
}
