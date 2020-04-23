package utils

import (
	"fmt"
	"os"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

const (
	colorBlack = iota + 30
	colorRed
	colorGreen
	colorYellow
	colorBlue
	colorMagenta
	colorCyan
	colorWhite

	colorBold     = 1
	colorDarkGray = 90
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

func colorize(s interface{}, c int, disabled bool) string {
	if disabled {
		return fmt.Sprintf("%s", s)
	}
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", c, s)
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
			noColor := c.Bool("log_nocolor")

			w.Out = f
			w.NoColor = noColor
			w.FormatTimestamp = emptyFormatter
			w.FormatLevel = func(level interface{}) string {
				ll, ok := level.(string)
				if !ok {
					if level == nil {
						return ""
					} else {
						return fmt.Sprint(level)
					}
				}

				switch ll {
				case "trace":
					return colorize("TRC", colorMagenta, noColor)
				case "debug":
					return colorize("DBG", colorYellow, noColor)
				case "info":
					return colorize("INF", colorGreen, noColor)
				case "warn":
					return colorize("WRN", colorRed, noColor)
				case "error":
					return colorize(colorize("ERR", colorRed, noColor), colorBold, noColor)
				case "fatal":
					return colorize(colorize("FTL", colorRed, noColor), colorBold, noColor)
				case "panic":
					return colorize(colorize("PNC", colorRed, noColor), colorBold, noColor)
				default:
					return colorize("???", colorBold, noColor)
				}
			}
		}))
	}

	return
}
