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

	_ "github.com/opencontainers/runc/libcontainer/nsenter"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
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

	flags := []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "/etc/chrootd/conf",
				EnvVars: []string{"CHROOTD_CONFIG"},
				Usage:   "load toml config from `FILE`",
			},
			altsrc.NewStringFlag(&cli.StringFlag{
				Name:    "addr",
				Usage:   "server connection addr",
				Value:   "127.0.0.1:9090",
				EnvVars: []string{"CHROOTD_CONNADDR"},
			}),
			altsrc.NewStringFlag(&cli.StringFlag{
				Name:    "network",
				Usage:   "server connection network type",
				Value:   "tcp",
				EnvVars: []string{"CHROOTD_CONNTYPE"},
			}),
			altsrc.NewDurationFlag(&cli.DurationFlag{
				Name:    "timeout",
				Usage:   "server connection timeout",
				Value:   10 * time.Second,
				EnvVars: []string{"CHROOTD_CONNTIMEOUT"},
			}),
			altsrc.NewStringFlag(&cli.StringFlag{
				Name:    "pid",
				Value:   "chrootd.pid",
				EnvVars: []string{"CHROOTD_PIDFILE"},
				Usage:   "daemon pid path",
			}),
			altsrc.NewStringFlag(&cli.StringFlag{
				Name:    "runpath",
				Value:   "./container",
				EnvVars: []string{"CHROOTD_RUN"},
				Usage:   "daemon run path",
			}),
			altsrc.NewStringFlag(&cli.StringFlag{
				Name:    "log",
				Value:   "chrootd.log",
				EnvVars: []string{"CHROOTD_LOGFILE"},
				Usage:   "daemon log path",
			}),
			altsrc.NewIntFlag(&cli.IntFlag{
				Name:  "loglevel",
				Value: 1,
				Usage: "set log level\n0 - debug\n1 - info\n2 - warn\n3 - error",
			}),
			altsrc.NewBoolFlag(&cli.BoolFlag{
				Name:  "daemon",
				Value: false,
				Usage: "start in background",
			}),
		}

	app := &cli.App{
		UseShortOptionHandling: true,
		Flags: flags,
		Before: altsrc.InitInputSourceWithContext(flags, altsrc.NewTomlSourceFromFlagFunc("config")),
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
			user.Timeout = c.Duration("timeout")
			user.PidFileName = c.String("pid")
			user.LogFileName = c.String("log")
			user.Network = c.String("network")
			user.Addr = c.String("addr")
			user.RunPath = c.String("runpath")

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
