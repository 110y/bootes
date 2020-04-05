package cache

import (
	"context"
	"fmt"

	apiv1 "github.com/110y/bootes/internal/k8s/api/v1"
	"github.com/110y/bootes/internal/observer/trace"
	xdscache "github.com/envoyproxy/go-control-plane/pkg/cache"
)

var _ Cache = (*cache)(nil)

type Cache interface {
	IsCachedNode(node string) bool
	UpdateAllResources(ctx context.Context, node, version string, clusters []*apiv1.Cluster, listeners []*apiv1.Listener) error
	UpdateClusters(ctx context.Context, node, version string, clusters []*apiv1.Cluster) error
	UpdateListeners(ctx context.Context, node, version string, listeners []*apiv1.Listener) error
	UpdateRoutes(ctx context.Context, node, version string, routes []*apiv1.Route) error
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

func (c *cache) UpdateAllResources(ctx context.Context, node, version string, clusters []*apiv1.Cluster, listeners []*apiv1.Listener) error {
	ctx, span := trace.NewSpan(ctx, "Cache.UpdateAllResources")
	defer span.End()

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

func (c *cache) UpdateClusters(ctx context.Context, node, version string, clusters []*apiv1.Cluster) error {
	ctx, span := trace.NewSpan(ctx, "Cache.UpdateClusters")
	defer span.End()

	snapshot := c.newClusterSnapshot(node, version, clusters)
	if err := c.snapshotCache.SetSnapshot(node, snapshot); err != nil {
		return fmt.Errorf("failed to update cluster snapshot: %w", err)
	}

	return nil
}

func (c *cache) UpdateListeners(ctx context.Context, node, version string, listeners []*apiv1.Listener) error {
	ctx, span := trace.NewSpan(ctx, "Cache.UpdateListeners")
	defer span.End()

	snapshot := c.newListenerSnapshot(node, version, listeners)
	if err := c.snapshotCache.SetSnapshot(node, snapshot); err != nil {
		return fmt.Errorf("failed to update listener snapshot: %w", err)
	}

	return nil
}

func (c *cache) UpdateRoutes(ctx context.Context, node, version string, routes []*apiv1.Route) error {
	ctx, span := trace.NewSpan(ctx, "Cache.UpdateRoutes")
	defer span.End()

	snapshot := c.newRouteSnapshot(node, version, routes)
	if err := c.snapshotCache.SetSnapshot(node, snapshot); err != nil {
		return fmt.Errorf("failed to update route snapshot: %w", err)
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

func (c *cache) newRouteSnapshot(node, version string, routes []*apiv1.Route) xdscache.Snapshot {
	resources := make([]xdscache.Resource, len(routes))
	for i, r := range routes {
		resources[i] = r.Spec.Config
	}

	s, err := c.snapshotCache.GetSnapshot(node)
	if err != nil {
		return xdscache.NewSnapshot(version, nil, nil, resources, nil, nil)
	}

	endpoints := getResourceFromSnapshot(&s, xdscache.EndpointType)
	clusters := getResourceFromSnapshot(&s, xdscache.ClusterType)
	listeners := getResourceFromSnapshot(&s, xdscache.ListenerType)
	runtimes := getResourceFromSnapshot(&s, xdscache.RuntimeType)

	return xdscache.NewSnapshot(version, endpoints, clusters, resources, listeners, runtimes)
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
