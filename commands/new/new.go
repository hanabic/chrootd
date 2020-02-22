package new

import (
	"flag"
	. "github.com/xhebox/chrootd/api/common"
	. "github.com/xhebox/chrootd/commands"
	"google.golang.org/grpc"
	"log"
)

var New = Command{
	Name: "new",
	Desc: "for test",
	Hanlder: func(args []string) error {
		fs := flag.NewFlagSet("new", flag.ContinueOnError)

		grpcConf := GrpcConfig{}
		grpcConf.SetFlag(fs)

		if err := fs.Parse(args); err != nil {
			return err
		}

		log.Printf("connecting to grpc server %s via %s\n", grpcConf.Url, grpcConf.NetWorkType)

		conn, err := grpcConf.Connect(grpc.WithInsecure())
		defer conn.Close()
		return err
	},
}
