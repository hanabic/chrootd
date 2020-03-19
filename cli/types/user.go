package types

import (
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

type User struct {
	Logger zerolog.Logger
	Conn   *grpc.ClientConn
}
