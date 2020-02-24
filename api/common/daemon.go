package common

import (
	"flag"
	"os"
)

type DaemonConfig struct {
	GrpcConn    ConnConfig
	ConfPath    string
	LogLevel    string
	PidFileName string
	PidFilePath string
	PidFilePerm int
	LogFileName string
	LogFilePerm int
	LogFilePath string
	WorkDir     string
	Chroot      string
}

func (conf *DaemonConfig) SetFlag(fs *flag.FlagSet) {
	conf.GrpcConn.SetFlag(fs)
	fs.StringVar(&conf.ConfPath, "p", "/etc/chrootd/daemon.ini", `set path of daemon config`)
}

func (conf *DaemonConfig) LoadEnv() {
	conf.GrpcConn.LoadEnv()

	if value := os.Getenv("CHROOTD_DAEMONCONFPATH"); value != "" {
		conf.ConfPath = value
	}
	if value := os.Getenv("CHROOTD_DAEMONPIDNAME"); value != "" {
		conf.PidFileName = value
	}
	if value := os.Getenv("CHROOTD_DAEMONPIDPATH"); value != "" {
		conf.PidFilePath = value
	}
	if value := os.Getenv("CHROOTD_DAEMONLOGNAME"); value != "" {
		conf.LogFileName = value
	}
	if value := os.Getenv("CHROOTD_DAEMONLOGPATH"); value != "" {
		conf.LogFilePath = value
	}
}
