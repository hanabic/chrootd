package main

import (
	"flag"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/sevlyar/go-daemon"
	. "github.com/xhebox/chrootd/api/containerpool/protobuf"
	. "github.com/xhebox/chrootd/api/containerpool/server"
	. "github.com/xhebox/chrootd/common"
	"google.golang.org/grpc"
)

var (
	signal *string
	stop   = make(chan struct{})
)

func termHandler(sig os.Signal) error {
	stop <- struct{}{}
	log.Println("terminate")
	return daemon.ErrStop
}

func main() {
	fs := flag.NewFlagSet("daemon", flag.ContinueOnError)

	signal = fs.String("s", "", `stop — shutdown`)

	daemonConf := NewDaemonConfig()
	daemonConf.GrpcConn.SetFlag(fs)

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatalln(err)
	}
	daemon.AddCommand(daemon.StringFlag(signal, "stop"), syscall.SIGINT, termHandler)

	daemonConf.LoadEnv()

	if err := daemonConf.ParseIni(); err != nil {
		log.Fatalln(err)
	}

	cntxt := &daemon.Context{
		PidFileName: daemonConf.PidFileName,
		PidFilePerm: daemonConf.PidFilePerm,
		LogFileName: daemonConf.LogFileName,
		LogFilePerm: daemonConf.LogFilePerm,
		WorkDir:     daemonConf.WorkDir,
		Umask:       027,
		Args:        []string{"[go-daemon sample]"},
	}

	if len(daemon.ActiveFlags()) > 0 {
		d, err := cntxt.Search()
		if err != nil {
			log.Fatalf("Unable send signal to the daemon: %v\n", err)
		}
		daemon.SendCommands(d)
		return
	}

	d, err := cntxt.Reborn()
	if err != nil {
		log.Fatalf("Unable to run: %v\n", err)
	}
	if d != nil {
		return
	}
	defer cntxt.Release()

	log.Println("daemon started")

	lis, err := daemonConf.GrpcConn.Listen()
	if err != nil {
		log.Fatalf("unable to listen: %v\n", err)
	}
	defer lis.Close()

	grpcServer := grpc.NewServer()
	pool := PoolServer{}
	pool.ContainerGroup = make(map[string]string)

	RegisterContainerPoolServer(grpcServer, &pool)
	log.Println("listening in ", daemonConf.GrpcConn.Addr, " ", daemonConf.GrpcConn.NetWorkType)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("grpc: failed to serve: %v\n", err)
			stop <- struct{}{}
		}
	}()

	go func() {
	loop:
		for {
			time.Sleep(time.Second)
			select {
			case <-stop:
				grpcServer.GracefulStop()
				break loop
			default:
			}
		}
	}()

	if err = daemon.ServeSignals(); err != nil {
		log.Printf("Fail to serve signals: %v\n", err)
	}
}
