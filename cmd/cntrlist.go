package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/urfave/cli/v2"
	mtyp "github.com/xhebox/chrootd/meta"
)

var CntrQuery = &cli.Command{
	Name:      "query",
	Usage:     "query all containers",
	ArgsUsage: "[ [$nodeid] a tidwall/gjson valid json query string ]",
	Aliases:   []string{"list"},
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		args := strings.Join(c.Args().Slice(), ",")
		if args == "" {
			args = "[]"
		}

		writer := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

		fmt.Fprintf(writer, "Name\tCntrId\tMetaID\tImage\n")

		err := user.Cntr.List(args, func(cntr string, meta *mtyp.Metainfo) error {
			fmt.Fprintf(writer, "%s\t%s\t%s\t%s:%s\n", meta.Name, cntr, meta.Id, meta.Image, meta.ImageReference)
			return nil
		})
		if err != nil {
			return err
		}

		writer.Flush()

		return nil
	},
}
