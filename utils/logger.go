package utils

import (
	"os"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

var (
	ZerologFlags = []cli.Flag{
		&cli.PathFlag{
			Name:  "log",
			Value: "-",
			Usage: "`LOGFILE` path, defaults to stdout",
		},
		&cli.StringFlag{
			Name:  "log_level",
			Value: "info",
			Usage: "set `loglevel`: debug, info, warn, error",
		},
		&cli.BoolFlag{
			Name:  "log_structure",
			Value: false,
			Usage: "output structed json log",
		},
		&cli.BoolFlag{
			Name:  "log_nocolor",
			Value: false,
			Usage: "when log is not structed, default with color",
		},
	}
)

func emptyFormatter(interface{}) string {
	return ""
}

func NewLogger(c *cli.Context, l zerolog.Logger) (r zerolog.Logger, err error) {
	r = l

	logPath := c.Path("log")
	var f *os.File
	if logPath == "-" {
		f = os.Stdout
	} else {
		f, err = os.OpenFile(c.Path("log"), os.O_CREATE|os.O_RDWR|syscall.O_CLOEXEC, 0644)
		if err != nil {
			return
		}
	}
	r = r.Output(f)

	switch c.String("log_level") {
	case "debug":
		r = r.Level(zerolog.DebugLevel)
	case "info":
		r = r.Level(zerolog.InfoLevel)
	case "warn":
		r = r.Level(zerolog.WarnLevel)
	case "error":
		r = r.Level(zerolog.ErrorLevel)
	}

	if c.Bool("log_structure") {
		r = r.With().Timestamp().Logger()
	} else {
		r = r.Output(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
			w.Out = f
			w.NoColor = c.Bool("log_nocolor")
			w.FormatTimestamp = emptyFormatter
		}))
	}

	return
}
