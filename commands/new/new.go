package new

import (
	"flag"
	"fmt"
	. "github.com/xhebox/chrootd/api/common"
	. "github.com/xhebox/chrootd/commands"
	"google.golang.org/grpc"
)

func Connect(conf *GrpcConfig, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.Dial(conf.Url, opts...)
}

var New = Command{
	Name: "new",
	Desc: "for test",
	Hanlder: func(args []string) error {
		var conf GrpcConfig
		fs := flag.NewFlagSet("new", flag.ContinueOnError)
		conf.Parse(fs)

		err := fs.Parse(args)
		if err != nil {
			return err
		}

		fmt.Printf("connecting to grpc server %s via %s\n", conf.Url, conf.NetWorkType)
		conn, err := Connect(&conf, grpc.WithInsecure())
		defer conn.Close()
		return err
	},
}
