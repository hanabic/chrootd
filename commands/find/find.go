package find

import (
	"context"
	"flag"
	"fmt"
	"github.com/xhebox/chrootd/api/containerpool"
	. "github.com/xhebox/chrootd/commands"
	. "github.com/xhebox/chrootd/common"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
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
		if err := fs.Parse(args); err != nil {
			return err
		}
		connConf.LoadEnv()

		conn, err := grpc.Dial("new", grpc.WithInsecure(), grpc.WithBlock(), grpc.WithContextDialer(func(ctx context.Context, target string) (net.Conn, error) {
			return connConf.PoolDial()
		}))
		if err != nil {
			return fmt.Errorf("did not connect: %v", err)
		}
		defer conn.Close()

		client := containerpool.NewContainerPoolClient(conn)

		reply, err := FindContainer(client, *name, *isCreate)
		if err != nil {
			return err
		}

		log.Printf("name: %s\t%s", *name, *reply)
		return nil
	},
}

func FindContainer(client containerpool.ContainerPoolClient, name string, isCreate bool) (*containerpool.Reply, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return client.FindContainer(ctx, &containerpool.Query{
		Name:     name,
		IsCreate: isCreate,
	})

}
