package utils

import (
	"flag"
	"fmt"
	"os"

	"github.com/docker/go-units"
	"github.com/urfave/cli/v2"
)

type SizeFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	EnvVars     []string
	Required    bool
	Hidden      bool
	Value       int64
	DefaultText string
	Destination *int64
	HasBeenSet  bool
}

func (f *SizeFlag) Apply(set *flag.FlagSet) error {
	for _, env := range f.EnvVars {
		envVar := os.Getenv(env)
		if envVar != "" {
			val, err := units.FromHumanSize(envVar)
			if err != nil {
				return fmt.Errorf("could not parse %q as bool value for flag %s: %s", val, f.Name, err)
			}

			f.Value = val
			f.HasBeenSet = true
		}
	}

	for _, name := range f.Names() {
		if f.Destination != nil {
			set.Int64Var(f.Destination, name, f.Value, f.Usage)
			continue
		}
		set.Int64(name, f.Value, f.Usage)
	}

	return nil
}

func (f *SizeFlag) String() string {
	return cli.FlagStringer(f)
}

func (f *SizeFlag) Names() []string {
	return append([]string{f.Name}, f.Aliases...)
}

func (f *SizeFlag) IsSet() bool {
	return f.HasBeenSet
}
