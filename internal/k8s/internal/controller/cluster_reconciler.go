package controller

import (
	"context"
	"fmt"

	"github.com/110y/bootes/internal/cache"
	api "github.com/110y/bootes/internal/k8s/api/v1"
	envoyapi "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	xdscache "github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/golang/protobuf/ptypes/duration"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterReconciler struct {
	client.Client
	Cache *cache.Cache
	// Scheme *runtime.Scheme
}

func (r *ClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	fmt.Println("START: RECONCILE")

	var cluster api.Cluster
	if err := r.Get(ctx, req.NamespacedName, &cluster); err != nil {
		// TODO:
		fmt.Println(fmt.Sprintf("RECONCILE: ERROR: %s", err))
		return ctrl.Result{}, err
	}

	fmt.Println(fmt.Sprintf("RECONCILE: %s", cluster.Spec.Name))

	var clusters []xdscache.Resource
	c := &envoyapi.Cluster{
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
	clusters = append(clusters, c)

	snapshot := xdscache.NewSnapshot("2.0)", nil, clusters, nil, nil, nil)
	if err := r.Cache.SetSnapshot("id", snapshot); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).For(&api.Cluster{}).Complete(r)
}
