package cache

import (
	"fmt"

	apiv1 "github.com/110y/bootes/internal/k8s/api/v1"
	xdscache "github.com/envoyproxy/go-control-plane/pkg/cache"
)

type Cache interface {
	UpdateClusters(node, version string, cluster []*apiv1.Cluster) error
}

type cache struct {
	snapshotCache xdscache.SnapshotCache
}

func New(snapshotCache xdscache.SnapshotCache) Cache {
	return &cache{
		snapshotCache: snapshotCache,
	}
}

func (c *cache) UpdateClusters(node, version string, clusters []*apiv1.Cluster) error {
	snapshot := c.newClusterSnapshot(node, version, clusters)
	if err := c.snapshotCache.SetSnapshot(node, snapshot); err != nil {
		return fmt.Errorf("failed to set snapshot: %w", err)
	}

	return nil
}

func (c *cache) newClusterSnapshot(node, version string, clusters []*apiv1.Cluster) xdscache.Snapshot {
	resources := make([]xdscache.Resource, len(clusters))
	for i, c := range clusters {
		resources[i] = c.Spec.Config
	}

	s, err := c.snapshotCache.GetSnapshot(node)
	if err != nil {
		return xdscache.NewSnapshot(version, nil, resources, nil, nil, nil)
	}

	endpoints := getResourceFromSnapshot(&s, xdscache.EndpointType)
	routes := getResourceFromSnapshot(&s, xdscache.RouteType)
	listeners := getResourceFromSnapshot(&s, xdscache.ListenerType)
	runtimes := getResourceFromSnapshot(&s, xdscache.RuntimeType)

	return xdscache.NewSnapshot(version, endpoints, resources, routes, listeners, runtimes)
}

func getResourceFromSnapshot(snapshot *xdscache.Snapshot, typeURL string) []xdscache.Resource {
	cache := snapshot.GetResources(typeURL)
	resources := make([]xdscache.Resource, len(cache))
	i := 0
	for _, e := range cache {
		resources[i] = e
		i++
	}

	return resources
}
