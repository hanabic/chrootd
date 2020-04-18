package utils

import (
	"github.com/rs/zerolog"
)

type rpcxLogger struct {
	zlog zerolog.Logger
}

func NewRpcxLogger(rs zerolog.Logger) *rpcxLogger {
	return &rpcxLogger{zlog: rs}
}

func (c *rpcxLogger) Debug(v ...interface{}) {
	c.zlog.Debug().Msgf("%+v", v)
}

func (c *rpcxLogger) Debugf(format string, v ...interface{}) {
	c.zlog.Debug().Msgf(format, v)
}

func (c *rpcxLogger) Info(v ...interface{}) {
	c.zlog.Info().Msgf("%+v", v)
}

func (c *rpcxLogger) Infof(format string, v ...interface{}) {
	c.zlog.Info().Msgf(format, v)
}

func (c *rpcxLogger) Warn(v ...interface{}) {
	c.zlog.Warn().Msgf("%+v", v)
}

func (c *rpcxLogger) Warnf(format string, v ...interface{}) {
	c.zlog.Warn().Msgf(format, v)
}

func (c *rpcxLogger) Error(v ...interface{}) {
	c.zlog.Error().Msgf("%+v", v)
}

func (c *rpcxLogger) Errorf(format string, v ...interface{}) {
	c.zlog.Error().Msgf(format, v)
}

func (c *rpcxLogger) Fatal(v ...interface{}) {
	c.zlog.Fatal().Msgf("%+v", v)
}

func (c *rpcxLogger) Fatalf(format string, v ...interface{}) {
	c.zlog.Fatal().Msgf(format, v)
}

func (c *rpcxLogger) Panic(v ...interface{}) {
	c.zlog.Panic().Msgf("%+v", v)
}

func (c *rpcxLogger) Panicf(format string, v ...interface{}) {
	c.zlog.Panic().Msgf(format, v)
}

