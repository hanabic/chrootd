package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/urfave/cli/v2"
)

var CntrGet = &cli.Command{
	Name:      "meta",
	Usage:     "get container metainfo",
	ArgsUsage: "$cntrid",
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		cntr, err := user.Cntr.Get(c.Args().First())
		if err != nil {
			return err
		}

		meta, err := cntr.Meta()
		if err != nil {
			return err
		}

		writer := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

		fmt.Fprintf(writer, "Name\tCntrId\tMetaID\tImage\n")

		fmt.Fprintf(writer, "%s\t%s\t%s\t%s:%s\n", meta.Name, c.Args().First(), meta.Id, meta.Image, meta.ImageReference)

		writer.Flush()

		return nil
	},
}
