package start

import (
	"context"
	"flag"
	"log"
	"net"
	"time"

	. "github.com/xhebox/chrootd/api/container"
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
		log.Println(connConf)

		log.Printf("connecting to grpc server %s via %s\n", connConf.Addr, connConf.NetWorkType)

		conn, err := grpc.Dial("start container", grpc.WithInsecure(), grpc.WithContextDialer(func(ctx context.Context, target string) (net.Conn, error) {
			return connConf.Dial()
		}))
		if err != nil {
			return err
		}

		defer conn.Close()

		client := NewContainerClient(conn)
		return StartContainer(client, args[0])
	},
}

func StartContainer(client ContainerClient, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.Handle(ctx)
	if err != nil {
		return err
	}
	log.Printf("start container %v\n", id)

	if err := stream.Send(&Packet{Payload: &Packet_Id{Id: id}}); err != nil {
		log.Println("error in send id packet")
		return err
	}

	if err := stream.Send(&Packet{Payload: &Packet_Start{Start: "start"}}); err != nil {
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
