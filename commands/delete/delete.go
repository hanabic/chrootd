package delete

import (
	"flag"
	"fmt"
	. "github.com/xhebox/chrootd/api/containerpool/client"
	. "github.com/xhebox/chrootd/api/containerpool/protobuf"
	. "github.com/xhebox/chrootd/commands"
	. "github.com/xhebox/chrootd/common"
	"google.golang.org/grpc"
	"log"
)

var Delete = Command{
	Name: "find",
	Desc: "find a container",
	Hanlder: func(args []string) error {

		fs := flag.NewFlagSet("delete", flag.ContinueOnError)
		id := fs.String("id", "hello", "id of container")
		connConf := ConnConfig{}
		connConf.SetFlag(fs)
		connConf.LoadEnv()

		if err := fs.Parse(args); err != nil {
			return err
		}

		conn, err := grpc.Dial(connConf.Addr, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		//conn, err := grpc.Dial(connConf.Addr, grpc.WithInsecure(), grpc.WithContextDialer(func(ctx context.Context, target string) (net.Conn, error) {
		//	return connConf.Dial()
		//}))
		defer conn.Close()

		if err != nil {
			return err
		}

		client := NewContainerPoolClient(conn)

		err = DeleteContainerById(client, *id)

		if err != nil {
			return err
		}
		fmt.Println("delete ", id, " successfully")
		return nil
	},
}
