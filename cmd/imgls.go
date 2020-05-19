package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/urfave/cli/v2"
)

var ImgList = &cli.Command{
	Name:      "list",
	Usage:     "list all available images on remote servers, or unpacked rootfs of given metadatas",
	Aliases:   []string{"ls", "l"},
	ArgsUsage: "[$meta1id, ..., $metaXid]",
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		writer := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

		if c.Args().Len() == 0 {
			fmt.Fprintf(writer, "NodeID\tName\tRefs\n")

			err := user.Meta.ImageAvailable(c.Context, func(id string, name string, refs []string) error {
				fmt.Fprintf(writer, "%s\t%s\t%v\n", id, name, refs)
				return nil
			})
			if err != nil {
				return err
			}
		} else {
			fmt.Fprintf(writer, "NodeID\tName\tRootfs\n")

			for _, v := range c.Args().Slice() {
				meta, err := user.Meta.Get(v)
				if err != nil {
					return err
				}

				fmt.Fprintf(writer, "%s\t%s\t%v\n", v, meta.Name, meta.RootfsIds)
			}
		}

		writer.Flush()

		return nil
	},
}
