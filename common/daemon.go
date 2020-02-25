package common

import (
	"flag"
	"os"

	"github.com/go-ini/ini"
)

type DaemonConfig struct {
	GrpcConn    ConnConfig
	ConfPath    string
	LogLevel    int
	PidFileName string
	PidFilePerm os.FileMode
	LogFileName string
	LogFilePerm os.FileMode
	WorkDir     string
}

func NewDaemonConfig() *DaemonConfig {
	return &DaemonConfig{
		ConfPath:    "/etc/chrootd/conf.ini",
		LogLevel:    5,
		PidFileName: "/run/chrootd.pid",
		PidFilePerm: os.FileMode(0644),
		LogFileName: "/var/log/chrootd.log",
		LogFilePerm: os.FileMode(0640),
		WorkDir:     "./",
	}
}

func (conf *DaemonConfig) SetFlag(fs *flag.FlagSet) {
	conf.GrpcConn.SetFlag(fs)
	fs.StringVar(&conf.ConfPath, "config", "/etc/chrootd/conf.ini", `set path of daemon config`)
	fs.StringVar(&conf.PidFileName, "pid", "chrootd.pid", `set path of pidfile`)
	fs.StringVar(&conf.LogFileName, "log", "chrootd.log", `set path of logfile`)
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
