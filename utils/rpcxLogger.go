package utils

import (
	"fmt"

	"github.com/rs/zerolog"
)

type rpcxLogger struct {
	zlog zerolog.Logger
}

func NewRpcxLogger(rs zerolog.Logger) *rpcxLogger {
	return &rpcxLogger{zlog: rs.With().Str("logger", "rpcx").Logger()}
}

func (c *rpcxLogger) Debug(v ...interface{}) {
	c.zlog.Debug().Msgf(fmt.Sprint(v...))
}

func (c *rpcxLogger) Debugf(format string, v ...interface{}) {
	c.zlog.Debug().Msgf(format, v...)
}

func (c *rpcxLogger) Info(v ...interface{}) {
	c.zlog.Info().Msgf(fmt.Sprint(v...))
}

func (c *rpcxLogger) Infof(format string, v ...interface{}) {
	c.zlog.Info().Msgf(format, v...)
}

func (c *rpcxLogger) Warn(v ...interface{}) {
	c.zlog.Warn().Msgf(fmt.Sprint(v...))
}

func (c *rpcxLogger) Warnf(format string, v ...interface{}) {
	c.zlog.Warn().Msgf(format, v...)
}

func (c *rpcxLogger) Error(v ...interface{}) {
	c.zlog.Error().Msgf(fmt.Sprint(v...))
}

func (c *rpcxLogger) Errorf(format string, v ...interface{}) {
	c.zlog.Error().Msgf(format, v...)
}

func (c *rpcxLogger) Fatal(v ...interface{}) {
	c.zlog.Fatal().Msgf(fmt.Sprint(v...))
}

func (c *rpcxLogger) Fatalf(format string, v ...interface{}) {
	c.zlog.Fatal().Msgf(format, v...)
}

func (c *rpcxLogger) Panic(v ...interface{}) {
	c.zlog.Panic().Msg(fmt.Sprint(v...))
}

func (c *rpcxLogger) Panicf(format string, v ...interface{}) {
	c.zlog.Panic().Msgf(format, v...)
}

