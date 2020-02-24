package commands

type Command struct {
	Name    string
	Desc    string
	Hanlder func([]string) error
}
