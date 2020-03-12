package cache

import (
	apiv1 "github.com/110y/bootes/internal/k8s/api/v1"
	envoyapi "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	xdscache "github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/golang/protobuf/ptypes/duration"
)

type Cache interface {
	UpdateClusters(node, version string, cluster []apiv1.Cluster) error
}

type cache struct {
	snapshotCache xdscache.SnapshotCache
}

func New(snapshotCache xdscache.SnapshotCache) Cache {
	return &cache{
		snapshotCache: snapshotCache,
	}
}

func (c *cache) UpdateClusters(node, version string, clusters []apiv1.Cluster) error {
	var resources []xdscache.Resource
	for _, cl := range clusters {
		resource := &envoyapi.Cluster{
			Name: cl.Spec.Name,
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
		resources = append(resources, resource)
	}

	// TODO: update only clusters by using GetSnapshot
	snapshot := xdscache.NewSnapshot(version, nil, resources, nil, nil, nil)
	if err := c.snapshotCache.SetSnapshot(node, snapshot); err != nil {
		return err
	}

	return nil
}
