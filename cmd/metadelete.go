package main

import (
	"regexp"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	mtyp "github.com/xhebox/chrootd/meta"
)

var MetaDelete = &cli.Command{
	Name:      "delete",
	Usage:     "delete metadata",
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

		for _, m := range res {
			err := user.Meta.Delete(m)
			if err != nil {
				return err
			}
		}

		return nil
	},
}
