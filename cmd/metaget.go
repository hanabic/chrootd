package main

import (
	"fmt"
	"os"
	"regexp"
	"text/tabwriter"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	mtyp "github.com/xhebox/chrootd/meta"
)

var MetaGet = &cli.Command{
	Name:      "get",
	Usage:     "output a container metadata to console",
	ArgsUsage: "[ [$metaid] [tidwall/gjson valid query string] ]",
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		res := []string{}
		switch c.Args().Len() {
		case 1:
			reg := regexp.MustCompile("[\\w\\d]+,[\\w\\d]+")

			first := c.Args().First()

			if reg.MatchString(first) {
				res = append(res, first)
				break
			}

			err := user.Meta.Query(first, func(meta *mtyp.Metainfo) error {
				res = append(res, meta.Id)
				return nil
			})
			if err != nil {
				return err
			}
		default:
			return errors.New("must specify at least one argument")
		}

		writer := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

		fmt.Fprintf(writer, "Name\tMetaID\tImage\n")

		for _, m := range res {
			meta, err := user.Meta.Get(m)
			if err != nil {
				return err
			}

			fmt.Fprintf(writer, "%s\t%s\t%s:%s\n", meta.Name, meta.Id, meta.Image, meta.ImageReference)
		}

		writer.Flush()

		return nil
	},
}
