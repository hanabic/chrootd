package find

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

var Find = Command{
	Name: "find",
	Desc: "find a container",
	Hanlder: func(args []string) error {

		fs := flag.NewFlagSet("new", flag.ContinueOnError)
		name := fs.String("name", "hello", "name of container")
		isCreate := fs.Bool("create", false, "whether to create")
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

		id, err := FindContainer(client, *name, *isCreate)

		if err != nil {
			return err
		}
		fmt.Println("name:", *name)
		fmt.Println("id:", id)
		return nil
	},
}
