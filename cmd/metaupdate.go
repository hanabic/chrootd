package main

import (
	"github.com/urfave/cli/v2"
	"github.com/xhebox/chrootd/utils"
)

var MetaUpdate = &cli.Command{
	Name:  "update",
	Usage: "update a container metadata",
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

		err = user.Meta.Update(meta)
		if err != nil {
			return err
		}

		return nil
	},
}
