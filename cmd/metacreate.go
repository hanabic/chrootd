package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/utils"
)

var MetaCreate = &cli.Command{
	Name:  "create",
	Usage: "create a container metadata",
	Flags: utils.ConcatMultipleFlags([]cli.Flag{},
		metaFlags,
		resourceFlags,
		capFlags,
	),
	Action: func(c *cli.Context) error {
		user := c.Context.Value("_data").(*User)

		meta, err := MetaFromCli(c)
		if err != nil {
			return err
		}

		id, err := user.Meta.Create(meta)
		if err != nil {
			return err
		}

		fmt.Printf("MetaID is %s\n", id)

		return nil
	},
}
