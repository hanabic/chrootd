package delete

import (
	"flag"
	. "github.com/xhebox/chrootd/api/containerpool/client"
	. "github.com/xhebox/chrootd/api/containerpool/protobuf"
	. "github.com/xhebox/chrootd/commands"
	. "github.com/xhebox/chrootd/common"
	"google.golang.org/grpc"
	"log"
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

		conn, err := grpc.Dial(connConf.Addr, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()

		client := NewContainerPoolClient(conn)

		if err := DeleteContainerById(client, *id); err != nil {
			return err
		}

		log.Println("delete ", id, " successfully")
		return nil
	},
}
