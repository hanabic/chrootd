package common

import (
	"flag"
	"os"

	"github.com/go-ini/ini"
)

type DaemonConfig struct {
	GrpcConn    ConnConfig
	ConfPath    string
	LogLevel    string
	PidFileName string
	PidFilePerm os.FileMode
	LogFileName string
	LogFilePerm os.FileMode
	WorkDir     string
	Chroot      string
}

func (conf *DaemonConfig) SetFlag(fs *flag.FlagSet) {
	conf.GrpcConn.SetFlag(fs)

	fs.StringVar(&conf.ConfPath, "p", "/etc/chrootd/conf.ini", `set path of daemon config`)

	conf.LogLevel = "file"
	conf.PidFileName = "/run/chrootd.pid"
	conf.LogFilePerm = os.FileMode(0644)
	conf.LogFileName = "/var/log/chrootd.log"
	conf.LogFilePerm = os.FileMode(0640)
	conf.WorkDir = "./"
	conf.Chroot = "/"
}

func (conf *DaemonConfig) LoadEnv() {
	conf.GrpcConn.LoadEnv()

	if value := os.Getenv("CHROOTD_CONFIG"); value != "" {
		conf.ConfPath = value
	}

	if value := os.Getenv("CHROOTD_PIDFILE"); value != "" {
		conf.PidFileName = value
	}

	if value := os.Getenv("CHROOTD_LOGFILE"); value != "" {
		conf.LogFileName = value
	}
}

func (conf *DaemonConfig) ParseIni() error {
	file, err := ini.Load(conf.ConfPath)
	if err != nil {
		return err
	}
	if err := file.Section("DAEMON").MapTo(conf); err != nil {
		return err
	}
	return nil
}

func (conf *DaemonConfig) Check() error {
	return nil
}
