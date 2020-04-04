package task

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/api"
	. "github.com/xhebox/chrootd/cli/types"
)

var Exec = &cli.Command{
	Name:  "exec",
	Usage: "exec a program in the specific task",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "id",
			Usage:    "task id",
			Required: true,
		},
		&cli.BoolFlag{
			Name:  "attach",
			Value: true,
			Usage: "redirect stdin/out/err to remote process",
		},
	},
	Action: func(c *cli.Context) error {
		data := c.Context.Value("_data").(*User)

		client := api.NewTaskClient(data.Conn)

		id, err := ksuid.Parse(c.String("id"))
		if err != nil {
			return err
		}

		args := []string{}
		if str := c.Args().First(); len(str) > 0 {
			args = append(args, str)
		} else {
			args = append(args, "/bin/ls")
		}
		args = append(args, c.Args().Tail()...)

		proc := &api.Proc{
			Args: args,
			Cwd: "/",
			ConsoleHeight: 64,
			ConsoleWidth: 80,
		}

		res, err := client.Exec(c.Context, &api.ExecReq{
			Id:     id.Bytes(),
			Attach: c.Bool("attach"),
			Prog:   proc,
		})
		if err != nil {
			return fmt.Errorf("fail to exec in task[%s]: %s", id, err)
		}
		if len(res.Reason) != 0 {
			return fmt.Errorf("fail to exec in task[%s]: %s", id, res.Reason)
		}

		info := res.Info

		if !c.Bool("attach") {
			// TODO: handwrittent proc info dump
			data.Logger.Info().Msgf("exec task[%s]: %v", id, proc)
			return nil
		}

		if info.Pid == -1 {
			return fmt.Errorf("invalid pid, can not attach")
		}

		handle := func(ctx context.Context, client api.TaskClient, id []byte, pid int64, typ string, w io.Writer) error {
			for {
				res, e := client.Read(ctx, &api.ReadReq{
					Id:   id,
					Pid:  pid,
					Type: typ,
				})
				if e != nil {
					if e == io.EOF {
						return nil
					}
					return e
				}
				if len(res.Reason) > 0 {
					if res.Reason == "eof" {
						return nil
					}
					return errors.New(res.Reason)
				}

				_, e = io.Copy(w, bytes.NewReader(res.D))
				if e != nil {
					return e
				}

				time.Sleep(200 * time.Microsecond)
			}
		}

		ch := make(chan bool)
		var err1, err2 error

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			err1 = handle(c.Context, client, id.Bytes(), info.Pid, "stdout", os.Stdout)
		}()
		go func() {
			defer wg.Done()
			err2 = handle(c.Context, client, id.Bytes(), info.Pid, "stderr", os.Stderr)
		}()
		go func() {
			wg.Wait()
			ch <- true
		}()

		go func() {
			buf := make([]byte, 512)
			for {
				n, err := os.Stdin.Read(buf)
				if err != nil {
					if err == io.EOF {
						return
					}
					// TODO: log error
					_ = err
				}

				res, err := client.Write(c.Context, &api.WriteReq{
					Id:  id.Bytes(),
					Pid: info.Pid,
					D:   buf[:n],
				})
				if err != nil {
					// TODO: log error
					_ = err
				}
				if len(res.Reason) > 0 {
					// TODO: log error
					_ = errors.New(res.Reason)
				}
			}
		}()

	loop:
		for {
			select {
			case <-ch:
				// TODO: log error
				_ = err1
				_ = err2
				break loop
			case <-c.Context.Done():
				break loop
			default:
				time.Sleep(100 * time.Microsecond)
			}
		}

		return nil
	},
}
