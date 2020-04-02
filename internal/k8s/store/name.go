package store

import (
	"fmt"
	"strings"
)

const nodeNameSeparator = "."

func ToNodeName(name, namespace string) string {
	return fmt.Sprintf("%s%s%s", name, nodeNameSeparator, namespace)
}

func ToNamespacedName(node string) (string, string) {
	parts := strings.SplitN(node, nodeNameSeparator, 2)
	if len(parts) == 1 {
		return parts[0], ""
	}

	return parts[0], parts[1]
}
