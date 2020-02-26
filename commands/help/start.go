package help

import (
	. "github.com/xhebox/chrootd/commands"
	"log"
)

var Help = Command{
	Name: "help",
	Desc: "help a container",
	Hanlder: func(args []string) error {
		log.Println("help!")
		return nil
	},
}
