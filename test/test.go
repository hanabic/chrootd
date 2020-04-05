package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
)

func main() {
	flags := []cli.Flag{
		altsrc.NewIntFlag(&cli.IntFlag{Name: "test"}),
		&cli.StringFlag{
			Name: "load",
			Value:"config.toml",
		},

	}

	app := &cli.App{
		Action: func(c *cli.Context) error {
			fmt.Println(c.Int("test"))
			return nil
		},
		Before: altsrc.InitInputSourceWithContext(flags, altsrc.NewTomlSourceFromFlagFunc("load")),
		Flags: flags,
	}

	app.Run(os.Args)
}