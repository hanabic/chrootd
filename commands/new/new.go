package new

import (
	"context"
	"flag"
	"log"
	"net"

	. "github.com/xhebox/chrootd/commands"
	. "github.com/xhebox/chrootd/common"
	"google.golang.org/grpc"
)

var New = Command{
	Name: "new",
	Desc: "for test",
	Hanlder: func(args []string) error {
		fs := flag.NewFlagSet("new", flag.ContinueOnError)

		connConf := ConnConfig{}
		connConf.SetFlag(fs)
		connConf.LoadEnv()

		if err := fs.Parse(args); err != nil {
			return err
		}

		log.Printf("connecting to grpc server %s via %s\n", connConf.PoolAddr, connConf.NetWorkType)

		conn, err := grpc.Dial("new", grpc.WithInsecure(), grpc.WithContextDialer(func(ctx context.Context, target string) (net.Conn, error) {
			return connConf.PoolDial()
		}))
		defer conn.Close()
		return err
	},
}
