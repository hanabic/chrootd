package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/urfave/cli/v2"
	mtyp "github.com/xhebox/chrootd/meta"
)

var MetaQuery = &cli.Command{
	Name:      "list",
	Usage:     "query metadatas",
	ArgsUsage: "[a tidwall/gjson valid json query string]",
	Aliases:   []string{"ls", "l"},
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		query := ""
		if c.Args().Len() > 0 {
			query = c.Args().First()
		}

		writer := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

		fmt.Fprintf(writer, "Name\tMetaID\tImage\n")

		err := user.Meta.Query(query, func(meta *mtyp.Metainfo) error {
			fmt.Fprintf(writer, "%s\t%s\t%s:%s\n", meta.Name, meta.Id, meta.Image, meta.ImageReference)
			return nil
		})
		if err != nil {
			return err
		}

		writer.Flush()

		return nil
	},
}
