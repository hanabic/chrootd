package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/urfave/cli/v2"
	ctyp "github.com/xhebox/chrootd/cntr"
)

var CntrQuery = &cli.Command{
	Name:      "list",
	Usage:     "query all containers",
	ArgsUsage: "[tidwall/gjson query string | container tag]",
	Aliases:   []string{"ls", "l"},
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		args := "[]"
		if c.Args().Len() == 1 {
			args = c.Args().First()
		} else if c.Args().Len() > 1 {
			args = strings.Join(c.Args().Slice(), ",")
		}

		writer := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

		fmt.Fprintf(writer, "Name\tTags\tRootfs\tCntrId\tMetaID\tImage\n")

		err := user.Cntr.List(args, func(info *ctyp.Cntrinfo) error {
			meta := info.Meta
			fmt.Fprintf(writer, "%s\t%v\t%s\t%s\t%s\t%s:%s\n", meta.Name, info.Tags, info.Rootfs, info.Id, meta.Id, meta.Image, meta.ImageReference)
			return nil
		})
		if err != nil {
			return err
		}

		writer.Flush()

		return nil
	},
}
