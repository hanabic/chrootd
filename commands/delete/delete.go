package delete

import (
	"context"
	"flag"
	"github.com/xhebox/chrootd/api/containerpool"
	. "github.com/xhebox/chrootd/commands"
	. "github.com/xhebox/chrootd/common"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

var Delete = Command{
	Name: "delete",
	Desc: "delete a container",
	Hanlder: func(args []string) error {
		fs := flag.NewFlagSet("delete", flag.ContinueOnError)
		id := fs.String("id", "hello", "id of container")
		connConf := ConnConfig{}
		connConf.SetFlag(fs)
		if err := fs.Parse(args); err != nil {
			return err
		}
		connConf.LoadEnv()

		conn, err := grpc.Dial("new", grpc.WithInsecure(), grpc.WithBlock(), grpc.WithContextDialer(func(ctx context.Context, target string) (net.Conn, error) {
			return connConf.Dial()
		}))
		if err != nil {
			log.Printf("did not connect: %v", err)
			return err
		}
		defer conn.Close()

		client := containerpool.NewContainerPoolClient(conn)

		reply, err := DeleteContainerById(client, *id)

		if err != nil {
			log.Printf("failed to delete: %v", err)
			return err
		}

		log.Printf("id:%v  %v", *id, *reply)

		return nil
	},
}

func DeleteContainerById(client containerpool.ContainerPoolClient, id string) (*containerpool.Reply, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return client.SetContainer(ctx, &containerpool.SetRequest{
		State: containerpool.StateType_Delete,
		Body: &containerpool.SetRequest_Delete{&containerpool.DeleteContainer{
			Id: id,
		}},
	})
}
