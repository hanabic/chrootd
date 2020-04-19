package utils

import (
	"github.com/urfave/cli/v2"
)

func ConcatMultipleFlags(flags ...[]cli.Flag) []cli.Flag {
	var l int
	for _, s := range flags {
		l += len(s)
	}
	tmp := make([]cli.Flag, l)
	var i int
	for _, s := range flags {
		i += copy(tmp[i:], s)
	}
	return tmp
}
