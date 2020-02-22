package main

import (
	"flag"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/sevlyar/go-daemon"
	. "github.com/xhebox/chrootd/api/common"
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

	connConf := ConnConfig{}
	connConf.SetFlag(fs)

	if err := fs.Parse(os.Args[1:]); err != nil {
		return
	}

	daemon.AddCommand(daemon.StringFlag(signal, "stop"), syscall.SIGINT, termHandler)

	cntxt := &daemon.Context{
		PidFileName: "chrootd.pid",
		PidFilePerm: 0644,
		LogFileName: "chrootd.log",
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
		Args:        []string{"[go-daemon sample]"},
	}

	if len(daemon.ActiveFlags()) > 0 {
		d, err := cntxt.Search()
		if err != nil {
			log.Fatalf("Unable send signal to the daemon: %s", err.Error())
		}
		daemon.SendCommands(d)
		return
	}

	d, err := cntxt.Reborn()
	if err != nil {
		log.Fatal("Unable to run: ", err)
	}
	if d != nil {
		return
	}
	defer cntxt.Release()

	log.Println("daemon started")

	lis, err := connConf.Listen()
	if err != nil {
		log.Fatal("unable to listen: ", err)
	}
	defer lis.Close()

	grpcServer := grpc.NewServer()

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("grpc: failed to serve: %v\n", err)
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

	err = daemon.ServeSignals()
	if err != nil {
		log.Printf("Error: %s", err.Error())
	}
}
