package main

import (
	"log"
	"os"

	. "github.com/xhebox/chrootd/commands"
	CommandDelete "github.com/xhebox/chrootd/commands/delete"
	CommandFind "github.com/xhebox/chrootd/commands/find"
	CommandHelp "github.com/xhebox/chrootd/commands/help"
	CommandNew "github.com/xhebox/chrootd/commands/new"
)

func main() {
	cmd := ""
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	commands := map[string]Command{
		"help":   CommandHelp.Help,
		"new":    CommandNew.New,
		"find":   CommandFind.Find,
		"delete": CommandDelete.Delete,
	}

	if v, ok := commands[cmd]; ok {
		if v.Hanlder != nil {
			args := []string{}
			if len(os.Args) > 2 {
				args = os.Args[2:]
			}

			if err := v.Hanlder(args); err != nil {
				log.Printf("%+v\n", err)
				return
			}
		}
	} else {
		log.Printf("%s\n", os.Args[0])
		for _, v := range commands {
			log.Printf("\t%s: %s\n", v.Name, v.Desc)
		}
	}
}
