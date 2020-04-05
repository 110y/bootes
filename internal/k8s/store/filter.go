package store

import (
	api "github.com/110y/bootes/internal/k8s/api/v1"
)

func FilterClustersByLabels(clusters []*api.Cluster, labels map[string]string) []*api.Cluster {
	results := []*api.Cluster{}
	for _, c := range clusters {
		if matchSelector(c, labels) {
			results = append(results, c)
		}
	}

	return results
}

func FilterListenersByLabels(listeners []*api.Listener, labels map[string]string) []*api.Listener {
	results := []*api.Listener{}
	for _, l := range listeners {
		if matchSelector(l, labels) {
			results = append(results, l)
		}
	}

	return results
}

func FilterRoutesByLabels(routes []*api.Route, labels map[string]string) []*api.Route {
	results := []*api.Route{}
	for _, r := range routes {
		if matchSelector(r, labels) {
			results = append(results, r)
		}
	}

	return results
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
