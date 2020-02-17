package k8s

import (
	"github.com/110y/bootes/internal/cache"
	"github.com/110y/bootes/internal/k8s/internal/controller"
	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	xdscache "github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/golang/protobuf/ptypes/duration"
)

type Controller struct {
	controller *controller.Controller
	cache      *cache.Cache
}

func NewController(sc *cache.Cache) (*Controller, error) {
	var clusters, endpoints, routes, listeners, runtimes []xdscache.Resource

	c := &api.Cluster{
		Name: "awesomecluster",
		ClusterDiscoveryType: &api.Cluster_Type{
			Type: api.Cluster_LOGICAL_DNS,
		},
		ConnectTimeout: &duration.Duration{Seconds: 1},
		LoadAssignment: &api.ClusterLoadAssignment{
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

	clusters = append(clusters, c)

	sp := xdscache.NewSnapshot("1.0)", endpoints, clusters, routes, listeners, runtimes)
	if err := sc.SetSnapshot("id", sp); err != nil {
		return nil, err
	}

	// TODO:
	ctrl, err := controller.NewController()
	if err != nil {
		// TODO:
		return nil, err
	}

	return &Controller{
		controller: ctrl,
		cache:      sc,
	}, nil
}

func (c *Controller) Start() error {
	return c.controller.Start()
}
