package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
)

var MetaGet = &cli.Command{
	Name:      "get",
	Usage:     "output container metadata to console",
	Aliases:   []string{"g"},
	ArgsUsage: "[$metaid1, ..., $metaidX]",
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		writer := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

		fmt.Fprintf(writer, "\tName\tMetaID\tImage\tRootfs\n")

		for k, m := range c.Args().Slice() {
			meta, err := user.Meta.Get(m)
			if err != nil {
				return err
			}

			fmt.Fprintf(writer, "%d\t%s\t%s\t%s:%s\t%s\n", k, meta.Name, meta.Id, meta.Image, meta.ImageReference, meta.RootfsIds)
			if user.Logger.GetLevel() >= zerolog.InfoLevel {
				fmt.Fprintf(writer, "\t%+v\n", meta)
			}
		}

		writer.Flush()

		return nil
	},
}
