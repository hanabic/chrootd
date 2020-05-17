package utils

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

func ComposeID(nodeid, id string) string {
	return fmt.Sprintf("%s,%s", nodeid, id)
}

func DecomposeID(id string) (string, string, error) {
	ids := strings.SplitN(id, ",", 2)
	if len(ids) != 2 {
		return "", "", errors.New("invalid id")
	}

	return ids[0], ids[1], nil
}
