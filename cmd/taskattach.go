package main

import (
	"bytes"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var TaskAttach = &cli.Command{
	Name:      "attach",
	Usage:     "attach a task",
	ArgsUsage: "$cntrid $taskid",
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		if c.Args().Len() < 2 {
			return errors.New("must specify at least two arguments")
		}

		args := c.Args().Slice()

		cntr, err := user.Cntr.Get(args[0])
		if err != nil {
			return err
		}

		rw, err := cntr.Attach(args[1])
		if err != nil {
			return err
		}
		defer rw.Close()

		ch := make(chan bool, 1)
		data := make(chan []byte, 1)
		h := make(chan os.Signal, 1)

		signal.Notify(h, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			io.Copy(os.Stdout, rw)
			ch <- true
		}()
		go func() {
			buf := make([]byte, 1024)

			for {
				n, err := os.Stdin.Read(buf)
				if err != nil {
					if err == io.EOF {
						return
					}
				}

				data <- buf[:n]
			}
		}()

	outer:
		for {
			select {
			case <-h:
				_, err = io.Copy(rw, bytes.NewReader([]byte{0x03}))
				if err != nil {
					return err
				}
			case <-ch:
				break outer
			case <-c.Context.Done():
				break outer
			case d := <-data:
				_, err = io.Copy(rw, bytes.NewReader(d))
				if err != nil {
					return err
				}
			default:
				time.Sleep(25 * time.Millisecond)
			}
		}

		return nil
	},
}
