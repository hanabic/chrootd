package delete

import (
	"context"
	"flag"
	. "github.com/xhebox/chrootd/api/containerpool/client"
	. "github.com/xhebox/chrootd/api/containerpool/protobuf"
	. "github.com/xhebox/chrootd/commands"
	. "github.com/xhebox/chrootd/common"
	"google.golang.org/grpc"
	"log"
	"net"
)

var Delete = Command{
	Name: "delete",
	Desc: "delete a container",
	Hanlder: func(args []string) error {
		fs := flag.NewFlagSet("delete", flag.ContinueOnError)
		id := fs.String("id", "hello", "id of container")
		connConf := ConnConfig{}
		connConf.SetFlag(fs)
		connConf.LoadEnv()

		if err := fs.Parse(args); err != nil {
			return err
		}

		conn, err := grpc.Dial("new", grpc.WithInsecure(), grpc.WithBlock(), grpc.WithContextDialer(func(ctx context.Context, target string) (net.Conn, error) {
			return connConf.Dial()
		}))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()

		client := NewContainerPoolClient(conn)

		reply, err := DeleteContainerById(client, *id)

		if err != nil {
			log.Fatalf("failed to delete: %v", err)
		}

		log.Printf("id:%v  %v", *id, reply)

		return nil
	},
}
