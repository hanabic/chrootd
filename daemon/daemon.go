package main

import (
	"flag"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/sevlyar/go-daemon"
)

var (
	signal = flag.String("s", "", `
	stop — shutdown`)
)

var (
	stop = make(chan struct{})
)

func termHandler(sig os.Signal) error {
	stop <- struct{}{}
	log.Println("terminate")
	return daemon.ErrStop
}

func main() {
	flag.Parse()
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

	log.Print("daemon started")

	go func() {
	loop:
		for {
			time.Sleep(time.Second)
			select {
			case <-stop:
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
