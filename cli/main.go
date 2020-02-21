package main

import (
	"fmt"
	. "github.com/xhebox/chrootd/commands"
	CommandHelp "github.com/xhebox/chrootd/commands/help"
	CommandNew "github.com/xhebox/chrootd/commands/new"
	"os"
)

func main() {
	cmd := ""
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	commands := map[string]Command{
		"help": CommandHelp.Help,
		"new":  CommandNew.New,
	}

	if v, ok := commands[cmd]; ok {
		if v.Hanlder != nil {
			args := []string{}
			if len(os.Args) > 2 {
				args = os.Args[2:]
			}

			if err := v.Hanlder(args); err != nil {
				fmt.Printf("%+v\n", err)
				return
			}
		}
	} else {
		fmt.Printf("%s\n", os.Args[0])
		for _, v := range commands {
			fmt.Printf("\t%s: %s\n", v.Name, v.Desc)
		}
	}
}
