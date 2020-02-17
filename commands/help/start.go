package help

import (
	"fmt"
	. "github.com/xhebox/chrootd/commands"
)

var Help = Command{
	Name: "help",
	Desc: "help a container",
	Hanlder: func(args []string) error {
		fmt.Println("help!")
		return nil
	},
}
