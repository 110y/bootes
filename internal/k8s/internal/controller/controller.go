package controller

import (
	"fmt"
)

const nodeNameSeparator = '.'

func toNodeName(name, namespace string) string {
	return fmt.Sprintf("%s%c%s", name, nodeNameSeparator, namespace)
}
