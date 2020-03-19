package utils

import (
	"os"
)

func PathExist(path string) bool {
	_, e := os.Stat(path)
	return !os.IsNotExist(e)
}
