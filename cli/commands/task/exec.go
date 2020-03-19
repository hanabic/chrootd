package task

import (
	"bytes"
	"fmt"
	"io"
	"os"

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
			Name:  "init",
			Value: true,
			Usage: "if it is the init process",
		},
		&cli.BoolFlag{
			Name:  "attach",
			Value: true,
			Usage: "redirect stdin/out/err to remote process",
		},
		&cli.StringSliceFlag{
			Name:  "args",
			Value: cli.NewStringSlice("/bin/ls"),
			Usage: "program args, args[0] is the execution binary",
		},
	},
	Action: func(c *cli.Context) error {
		data := c.Context.Value("_data").(*User)

		client := api.NewTaskClient(data.Conn)

		id, err := ksuid.Parse(c.String("id"))
		if err != nil {
			return err
		}

		proc := &api.Proc{
			Args: c.StringSlice("args"),
			Init: c.Bool("init"),
		}

		res, err := client.Exec(c.Context, &api.ExecReq{
			Id:   id.Bytes(),
			Attach: c.Bool("attach"),
			Prog: proc,
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

		tun, err := client.IO(c.Context)
		if err != nil {
			return errors.Wrapf(err, "fail to grab process %d of task[%s]", info.Pid, id)
		}

		err = tun.Send(&api.IOReq{
			Id:  id.Bytes(),
			Pid: info.Pid,
		})
		if err != nil {
			return errors.Wrapf(err, "fail to grab process %d of task[%s]", info.Pid, id)
		}

		first, err := tun.Recv()
		if err != nil || first.Str != "handshaked" {
			return errors.Wrapf(err, "fail to grab process %d of task[%s]", info.Pid, id)
		}

		handle := func(file *os.File, cat string, tun api.Task_IOClient) error {
		loop:
			for {
				select {
				case <-tun.Context().Done():
					break loop
				default:
					res, e := tun.Recv()
					if e != nil {
						if e == io.EOF {
							break loop
						}
						return e
					}

					if res.Str != cat {
						return errors.New("different cat, stop")
					}

					_, e = io.Copy(file, bytes.NewReader(res.D))
					if e != nil {
						return e
					}
				}
			}
			return nil
		}

		go handle(os.Stdout, "stdout", tun)
		go handle(os.Stderr, "stderr", tun)

		buf := make([]byte, 512)
	loop:
		for {
			select {
			case <-tun.Context().Done():
				break loop
			default:
				n, err := os.Stdin.Read(buf)
				if err != nil {
					if err == io.EOF {
						break loop
					}
					return err
				}

				if err := tun.Send(&api.IOReq{
					D: buf[:n],
				}); err != nil {
					return err
				}
			}
		}

		return nil
	},
}
