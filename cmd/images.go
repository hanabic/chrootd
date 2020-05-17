package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/urfave/cli/v2"
)

var MetaIMGList = &cli.Command{
	Name:  "images",
	Usage: "list all available images on remote servers",
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		writer := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

		fmt.Fprintf(writer, "NodeID\tName\tRefs\n")

		err := user.Meta.ImageAvailable(c.Context, func(id string, name string, refs []string) error {
			fmt.Fprintf(writer, "%s\t%s\t%v\n", id, name, refs)
			return nil
		})
		if err != nil {
			return err
		}

		writer.Flush()

		return nil
	},
}
