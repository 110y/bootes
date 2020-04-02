package cache

import (
	"fmt"

	apiv1 "github.com/110y/bootes/internal/k8s/api/v1"
	xdscache "github.com/envoyproxy/go-control-plane/pkg/cache"
)

var _ Cache = (*cache)(nil)

type Cache interface {
	IsCachedNode(node string) bool
	UpdateAllResources(node, version string, clusters []*apiv1.Cluster, listeners []*apiv1.Listener) error
	UpdateClusters(node, version string, clusters []*apiv1.Cluster) error
	UpdateListeners(node, version string, listeners []*apiv1.Listener) error
}

type cache struct {
	snapshotCache xdscache.SnapshotCache
}

func New(snapshotCache xdscache.SnapshotCache) Cache {
	return &cache{
		snapshotCache: snapshotCache,
	}
}

func (c *cache) IsCachedNode(node string) bool {
	_, err := c.snapshotCache.GetSnapshot(node)
	if err != nil {
		return false
	}

	return true
}

func (c *cache) UpdateAllResources(node, version string, clusters []*apiv1.Cluster, listeners []*apiv1.Listener) error {
	cr := make([]xdscache.Resource, len(clusters))
	for i, c := range clusters {
		cr[i] = c.Spec.Config
	}

	lr := make([]xdscache.Resource, len(listeners))
	for i, l := range listeners {
		lr[i] = l.Spec.Config
	}

	var s xdscache.Snapshot
	oldSnapshot, err := c.snapshotCache.GetSnapshot(node)
	if err != nil {
		s = xdscache.NewSnapshot(version, nil, cr, nil, lr, nil)
	} else {
		endpoints := getResourceFromSnapshot(&oldSnapshot, xdscache.EndpointType)
		routes := getResourceFromSnapshot(&oldSnapshot, xdscache.RouteType)
		runtimes := getResourceFromSnapshot(&oldSnapshot, xdscache.RuntimeType)

		xdscache.NewSnapshot(version, endpoints, cr, routes, lr, runtimes)
	}

	if err := c.snapshotCache.SetSnapshot(node, s); err != nil {
		return fmt.Errorf("failed to update all resources snapshot: %w", err)
	}

	return nil
}

func (c *cache) UpdateClusters(node, version string, clusters []*apiv1.Cluster) error {
	snapshot := c.newClusterSnapshot(node, version, clusters)
	if err := c.snapshotCache.SetSnapshot(node, snapshot); err != nil {
		return fmt.Errorf("failed to update cluster snapshot: %w", err)
	}

	return nil
}

func (c *cache) UpdateListeners(node, version string, listeners []*apiv1.Listener) error {
	snapshot := c.newListenerSnapshot(node, version, listeners)
	if err := c.snapshotCache.SetSnapshot(node, snapshot); err != nil {
		return fmt.Errorf("failed to update listener snapshot: %w", err)
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

func (c *cache) newListenerSnapshot(node, version string, listeners []*apiv1.Listener) xdscache.Snapshot {
	resources := make([]xdscache.Resource, len(listeners))
	for i, l := range listeners {
		resources[i] = l.Spec.Config
	}

	s, err := c.snapshotCache.GetSnapshot(node)
	if err != nil {
		return xdscache.NewSnapshot(version, nil, nil, nil, resources, nil)
	}

	endpoints := getResourceFromSnapshot(&s, xdscache.EndpointType)
	clusters := getResourceFromSnapshot(&s, xdscache.ClusterType)
	routes := getResourceFromSnapshot(&s, xdscache.RouteType)
	runtimes := getResourceFromSnapshot(&s, xdscache.RuntimeType)

	return xdscache.NewSnapshot(version, endpoints, clusters, routes, resources, runtimes)
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
