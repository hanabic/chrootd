package main

import (
	"context"
	"errors"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-ini/ini"
	_ "github.com/opencontainers/runc/libcontainer/nsenter"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/api"
	"google.golang.org/grpc"
)

var (
	stop           = make(chan struct{})
	ErrDaemonStart = errors.New("daemon start")
)

type User struct {
	Logger      zerolog.Logger
	Network     string
	Addr        string
	Timeout     time.Duration
	ConfPath    string
	RunPath     string
	PidFileName string
	PidFilePerm os.FileMode
	LogFileName string
	LogFilePerm os.FileMode
}

func main() {
	user := &User{
		Logger: zerolog.New(os.Stdout).With().Timestamp().Logger(),
	}

	app := &cli.App{
		UseShortOptionHandling: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "addr",
				Usage:   "server connection addr",
				Value:   "127.0.0.1:9090",
				EnvVars: []string{"CHROOTD_CONNADDR"},
			},
			&cli.StringFlag{
				Name:    "network",
				Usage:   "server connection network type",
				Value:   "tcp",
				EnvVars: []string{"CHROOTD_CONNTYPE"},
			},
			&cli.DurationFlag{
				Name:    "timeout",
				Usage:   "server connection timeout",
				Value:   10 * time.Second,
				EnvVars: []string{"CHROOTD_CONNTIMEOUT"},
			},
			&cli.StringFlag{
				Name:    "config",
				Value:   "/etc/chrootd/conf.ini",
				EnvVars: []string{"CHROOTD_CONFIG"},
				Usage:   "daemon config path",
			},
			&cli.StringFlag{
				Name:    "pid",
				Value:   "chrootd.pid",
				EnvVars: []string{"CHROOTD_PIDFILE"},
				Usage:   "daemon pid path",
			},
			&cli.StringFlag{
				Name:    "runpath",
				Value:   "./container",
				EnvVars: []string{"CHROOTD_RUN"},
				Usage:   "daemon run path",
			},
			&cli.StringFlag{
				Name:    "log",
				Value:   "chrootd.log",
				EnvVars: []string{"CHROOTD_LOGFILE"},
				Usage:   "daemon log path",
			},
			&cli.IntFlag{
				Name:  "loglevel",
				Value: 1,
				Usage: "set log level\n0 - debug\n1 - info\n2 - warn\n3 - error",
			},
			&cli.BoolFlag{
				Name:  "daemon",
				Value: false,
				Usage: "start in background",
			},
		},
		Action: func(c *cli.Context) error {
			user := c.Context.Value("_data").(*User)

			if c.Bool("daemon") {
				return ErrDaemonStart
			}

			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				<-sigs
				stop <- struct{}{}
			}()

			user.ConfPath = c.String("config")
			user.PidFileName = c.String("pid")
			user.LogFileName = c.String("log")

			file, err := ini.Load(user.ConfPath)
			if err != nil {
				return err
			}

			if err := file.Section("DAEMON").MapTo(user); err != nil {
				return err
			}

			user.Network = c.String("network")
			user.Addr = c.String("addr")
			user.Timeout = c.Duration("timeout")
			switch c.Int("loglevel") {
			case 0:
				user.Logger = user.Logger.Level(zerolog.DebugLevel)
			case 1:
				user.Logger = user.Logger.Level(zerolog.InfoLevel)
			case 2:
				user.Logger = user.Logger.Level(zerolog.WarnLevel)
			case 3:
				user.Logger = user.Logger.Level(zerolog.ErrorLevel)
			}
			user.RunPath = c.String("runpath")

			user.Logger.Log().Msgf("daemon started, logleve - %s", user.Logger.GetLevel())

			lis, err := net.Listen(user.Network, user.Addr)
			if err != nil {
				return err
			}
			defer lis.Close()

			cntrPool := newCntrPool()
			taskPool := newTaskPool()

			grpcServer := grpc.NewServer(grpc.ConnectionTimeout(user.Timeout))

			poolServer := newPoolServer(user, cntrPool)
			defer poolServer.Close()
			api.RegisterContainerPoolServer(grpcServer, poolServer)

			taskServer, err := newTaskServer(user, cntrPool, taskPool)
			if err != nil {
				return err
			}
			defer taskServer.Close()
			api.RegisterTaskServer(grpcServer, taskServer)

			user.Logger.Info().Msgf("listening server in [%s]%s", user.Network, user.Addr)
			go func() {
				if err := grpcServer.Serve(lis); err != nil {
					stop <- struct{}{}
					user.Logger.Error().Err(err).Msg("fail to serve")
				}
			}()

		loop:
			for {
				select {
				case <-stop:
					grpcServer.GracefulStop()
					break loop
				default:
					time.Sleep(500 * time.Microsecond)
				}
			}

			return nil
		},
	}

	ctx := context.WithValue(context.Background(), "_data", user)

	err := app.RunContext(ctx, os.Args)
	if err == ErrDaemonStart {
		skip := false
		args := make([]string, len(os.Args)-1)
		for _, v := range os.Args[1:] {
			switch {
			case v == "--daemon":
				skip = true
			case strings.HasPrefix(v, "--daemon=") || v == "-d" || skip:
				skip = false
			default:
				args = append(args, v)
			}
		}
		cmd := exec.Command(os.Args[0], args...)
		cmd.Start()
	} else if err != nil {
		user.Logger.Fatal().Msg(err.Error())
	}
}
