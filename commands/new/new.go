package new

import (
	"context"
	"flag"
	"log"
	"net"

	. "github.com/xhebox/chrootd/api/common"
	. "github.com/xhebox/chrootd/commands"
	"google.golang.org/grpc"
)

var New = Command{
	Name: "new",
	Desc: "for test",
	Hanlder: func(args []string) error {
		fs := flag.NewFlagSet("new", flag.ContinueOnError)

		connConf := ConnConfig{}
		connConf.SetFlag(fs)

		if err := fs.Parse(args); err != nil {
			return err
		}

		log.Printf("connecting to grpc server %s via %s\n", connConf.Url, connConf.NetWorkType)

		conn, err := grpc.Dial("new", grpc.WithInsecure(), grpc.WithContextDialer(func(ctx context.Context, target string) (net.Conn, error) {
			return connConf.Dial()
		}))
		defer conn.Close()
		return err
	},
}
