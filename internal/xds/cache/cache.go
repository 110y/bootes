package cache

import (
	"context"
	"fmt"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	xdscache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"

	apiv1 "github.com/110y/bootes-api/api/v1"
	"github.com/110y/bootes/internal/observer/trace"
)

var _ Cache = (*cache)(nil)

type Cache interface {
	IsCachedNode(node string) bool
	UpdateAllResources(ctx context.Context, node, version string, clusters []*apiv1.Cluster, listeners []*apiv1.Listener, routes []*apiv1.Route, endpoints []*apiv1.Endpoint) error
	UpdateClusters(ctx context.Context, node, version string, clusters []*apiv1.Cluster) error
	UpdateListeners(ctx context.Context, node, version string, listeners []*apiv1.Listener) error
	UpdateRoutes(ctx context.Context, node, version string, routes []*apiv1.Route) error
	UpdateEndpoints(ctx context.Context, node, version string, endpoints []*apiv1.Endpoint) error
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

func (c *cache) UpdateAllResources(ctx context.Context, node, version string, clusters []*apiv1.Cluster, listeners []*apiv1.Listener, routes []*apiv1.Route, endpoints []*apiv1.Endpoint) error {
	_, span := trace.NewSpan(ctx, "Cache.UpdateAllResources")
	defer span.End()

	cr := make([]types.Resource, len(clusters))
	for i, c := range clusters {
		cr[i] = c.Spec.Config
	}

	lr := make([]types.Resource, len(listeners))
	for i, l := range listeners {
		lr[i] = l.Spec.Config
	}

	rr := make([]types.Resource, len(routes))
	for i, r := range routes {
		rr[i] = r.Spec.Config
	}

	er := make([]types.Resource, len(endpoints))
	for i, e := range endpoints {
		er[i] = e.Spec.Config
	}

	var s xdscache.Snapshot
	oldSnapshot, err := c.snapshotCache.GetSnapshot(node)
	if err != nil {
		s = xdscache.NewSnapshot(version, er, cr, rr, lr, nil, nil)
	} else {
		runtimes := getResourceFromSnapshot(&oldSnapshot, resource.RuntimeType)
		secrets := getResourceFromSnapshot(&oldSnapshot, resource.SecretType)

		xdscache.NewSnapshot(version, er, cr, rr, lr, runtimes, secrets)
	}

	if err := c.snapshotCache.SetSnapshot(node, s); err != nil {
		return fmt.Errorf("failed to update all resources snapshot: %w", err)
	}

	return nil
}

func (c *cache) UpdateClusters(ctx context.Context, node, version string, clusters []*apiv1.Cluster) error {
	_, span := trace.NewSpan(ctx, "Cache.UpdateClusters")
	defer span.End()

	snapshot := c.newClusterSnapshot(node, version, clusters)
	if err := c.snapshotCache.SetSnapshot(node, snapshot); err != nil {
		return fmt.Errorf("failed to update cluster snapshot: %w", err)
	}

	return nil
}

func (c *cache) UpdateListeners(ctx context.Context, node, version string, listeners []*apiv1.Listener) error {
	_, span := trace.NewSpan(ctx, "Cache.UpdateListeners")
	defer span.End()

	snapshot := c.newListenerSnapshot(node, version, listeners)
	if err := c.snapshotCache.SetSnapshot(node, snapshot); err != nil {
		return fmt.Errorf("failed to update listener snapshot: %w", err)
	}

	return nil
}

func (c *cache) UpdateRoutes(ctx context.Context, node, version string, routes []*apiv1.Route) error {
	_, span := trace.NewSpan(ctx, "Cache.UpdateRoutes")
	defer span.End()

	snapshot := c.newRouteSnapshot(node, version, routes)
	if err := c.snapshotCache.SetSnapshot(node, snapshot); err != nil {
		return fmt.Errorf("failed to update route snapshot: %w", err)
	}

	return nil
}

func (c *cache) UpdateEndpoints(ctx context.Context, node, version string, endpoints []*apiv1.Endpoint) error {
	_, span := trace.NewSpan(ctx, "Cache.UpdateEndpoints")
	defer span.End()

	snapshot := c.newEndpointSnapshot(node, version, endpoints)
	if err := c.snapshotCache.SetSnapshot(node, snapshot); err != nil {
		return fmt.Errorf("failed to update endpoint snapshot: %w", err)
	}

	return nil
}

func (c *cache) newClusterSnapshot(node, version string, clusters []*apiv1.Cluster) xdscache.Snapshot {
	resources := make([]types.Resource, len(clusters))
	for i, c := range clusters {
		resources[i] = c.Spec.Config
	}

	s, err := c.snapshotCache.GetSnapshot(node)
	if err != nil {
		return xdscache.NewSnapshot(version, nil, resources, nil, nil, nil, nil)
	}

	endpoints := getResourceFromSnapshot(&s, resource.EndpointType)
	routes := getResourceFromSnapshot(&s, resource.RouteType)
	listeners := getResourceFromSnapshot(&s, resource.ListenerType)
	runtimes := getResourceFromSnapshot(&s, resource.RuntimeType)
	secrets := getResourceFromSnapshot(&s, resource.SecretType)

	return xdscache.NewSnapshot(version, endpoints, resources, routes, listeners, runtimes, secrets)
}

func (c *cache) newListenerSnapshot(node, version string, listeners []*apiv1.Listener) xdscache.Snapshot {
	resources := make([]types.Resource, len(listeners))
	for i, l := range listeners {
		resources[i] = l.Spec.Config
	}

	s, err := c.snapshotCache.GetSnapshot(node)
	if err != nil {
		return xdscache.NewSnapshot(version, nil, nil, nil, resources, nil, nil)
	}

	endpoints := getResourceFromSnapshot(&s, resource.EndpointType)
	clusters := getResourceFromSnapshot(&s, resource.ClusterType)
	routes := getResourceFromSnapshot(&s, resource.RouteType)
	runtimes := getResourceFromSnapshot(&s, resource.RuntimeType)
	secrets := getResourceFromSnapshot(&s, resource.SecretType)

	return xdscache.NewSnapshot(version, endpoints, clusters, routes, resources, runtimes, secrets)
}

func (c *cache) newRouteSnapshot(node, version string, routes []*apiv1.Route) xdscache.Snapshot {
	resources := make([]types.Resource, len(routes))
	for i, r := range routes {
		resources[i] = r.Spec.Config
	}

	s, err := c.snapshotCache.GetSnapshot(node)
	if err != nil {
		return xdscache.NewSnapshot(version, nil, nil, resources, nil, nil, nil)
	}

	endpoints := getResourceFromSnapshot(&s, resource.EndpointType)
	clusters := getResourceFromSnapshot(&s, resource.ClusterType)
	listeners := getResourceFromSnapshot(&s, resource.ListenerType)
	runtimes := getResourceFromSnapshot(&s, resource.RuntimeType)
	secrets := getResourceFromSnapshot(&s, resource.SecretType)

	return xdscache.NewSnapshot(version, endpoints, clusters, resources, listeners, runtimes, secrets)
}

func (c *cache) newEndpointSnapshot(node, version string, endpoints []*apiv1.Endpoint) xdscache.Snapshot {
	resources := make([]types.Resource, len(endpoints))
	for i, r := range endpoints {
		resources[i] = r.Spec.Config
	}

	s, err := c.snapshotCache.GetSnapshot(node)
	if err != nil {
		return xdscache.NewSnapshot(version, nil, nil, resources, nil, nil, nil)
	}

	clusters := getResourceFromSnapshot(&s, resource.ClusterType)
	listeners := getResourceFromSnapshot(&s, resource.ListenerType)
	routes := getResourceFromSnapshot(&s, resource.RouteType)
	runtimes := getResourceFromSnapshot(&s, resource.RuntimeType)
	secrets := getResourceFromSnapshot(&s, resource.SecretType)

	return xdscache.NewSnapshot(version, resources, clusters, routes, listeners, runtimes, secrets)
}

func getResourceFromSnapshot(snapshot *xdscache.Snapshot, typeURL string) []types.Resource {
	cache := snapshot.GetResources(typeURL)
	resources := make([]types.Resource, len(cache))
	i := 0
	for _, e := range cache {
		resources[i] = e
		i++
	}

	return resources
}
