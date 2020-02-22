package new

import (
	"flag"
	"fmt"
	. "github.com/xhebox/chrootd/api/common"
	. "github.com/xhebox/chrootd/commands"
	"google.golang.org/grpc"
)

var New = Command{
	Name: "new",
	Desc: "for test",
	Hanlder: func(args []string) error {
		var conf GrpcConfig
		fs := flag.NewFlagSet("new", flag.ContinueOnError)
		err := fs.Parse(args)
		if err != nil {
			return err
		}
		conf.Parse(fs)

		fmt.Printf("connecting to grpc server %s via %s\n", conf.Url, conf.NetWorkType)
		conn, err := conf.Connect(grpc.WithInsecure())
		defer conn.Close()
		return err
	},
}
