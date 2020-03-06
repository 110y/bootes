package cache

import (
	"sync"

	apiv1 "github.com/110y/bootes/internal/k8s/api/v1"
	envoyapi "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/golang/protobuf/ptypes/duration"
)

type xdsCache struct {
	clusters []*apiv1.Cluster
}

type Cache struct {
	cache.SnapshotCache

	mu        sync.Mutex
	nodeCache map[string]xdsCache
}

func NewCache() *Cache {
	sc := cache.NewSnapshotCache(false, cache.IDHash{}, nil)

	return &Cache{SnapshotCache: sc}
}

func (c *Cache) AddCluster(node string, cluster *apiv1.Cluster) error {
	// _, ok := c.nodeCache[node]
	// if !ok {
	//     // TODO: return error not found
	//     return nil
	// }

	cl := &envoyapi.Cluster{
		Name: cluster.Spec.Name,
		ClusterDiscoveryType: &envoyapi.Cluster_Type{
			Type: envoyapi.Cluster_LOGICAL_DNS,
		},
		ConnectTimeout: &duration.Duration{Seconds: 1},
		LoadAssignment: &envoyapi.ClusterLoadAssignment{
			ClusterName: "awesomeassign",
			Endpoints: []*endpoint.LocalityLbEndpoints{
				{
					LbEndpoints: []*endpoint.LbEndpoint{
						{
							HostIdentifier: &endpoint.LbEndpoint_Endpoint{
								Endpoint: &endpoint.Endpoint{
									Address: &core.Address{
										Address: &core.Address_SocketAddress{
											SocketAddress: &core.SocketAddress{
												Address: "foo",
												PortSpecifier: &core.SocketAddress_PortValue{
													PortValue: 9090,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	clusters := []cache.Resource{cl}

	snapshot := cache.NewSnapshot("2.0)", nil, clusters, nil, nil, nil)
	if err := c.SnapshotCache.SetSnapshot(node, snapshot); err != nil {
		return err
	}

	return nil
}
