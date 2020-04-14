package utils

import (
	"fmt"

	toml "github.com/pelletier/go-toml"
	"github.com/urfave/cli/v2"
)

func flagSet(tree *toml.Tree, v cli.Flag, c *cli.Context, config string) error {
	if v.IsSet() {
		return nil
	}

	key := v.Names()[0]
	if key == config {
		return nil
	}

	i := tree.Get(key)
	if i != nil {
		if err := c.Set(key, fmt.Sprint(i)); err != nil {
			return err
		}
	}

	return nil
}

func iterateCmds(cmd *cli.Command, tree *toml.Tree, c *cli.Context, config string) error {
	for _, v := range cmd.Flags {
		if err := flagSet(tree, v, c, config); err != nil {
			return err
		}
	}

	for _, cmd := range cmd.Subcommands {
		if err := iterateCmds(cmd, tree, c, config); err != nil {
			return err
		}
	}

	return nil
}

func NewTomlFlagLoader(config string) cli.BeforeFunc {
	return func(c *cli.Context) error {
		file := c.String(config)

		if file == "" {
			return nil
		}

		tree, err := toml.LoadFile(file)
		if err != nil {
			return err
		}

		for _, v := range c.App.Flags {
			if err := flagSet(tree, v, c, config); err != nil {
				return err
			}
		}

		for _, cmd := range c.App.Commands {
			if err := iterateCmds(cmd, tree, c, config); err != nil {
				return err
			}
		}

		return nil
	}
}
