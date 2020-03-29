package controller

import (
	"fmt"

	api "github.com/110y/bootes/internal/k8s/api/v1"
)

const nodeNameSeparator = '.'

func toNodeName(name, namespace string) string {
	return fmt.Sprintf("%s%c%s", name, nodeNameSeparator, namespace)
}

func matchSelector(resource api.EnvoyResource, labels map[string]string) bool {
	ws := resource.GetWorkloadSelector()
	if ws == nil {
		return true
	}

	match := true
	for key, val := range ws.Labels {
		v, ok := labels[key]
		if !ok || v != val {
			match = false
			break
		}
	}

	return match
}
