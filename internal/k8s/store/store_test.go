package store_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/golang/protobuf/proto"
	any "github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	api "github.com/110y/bootes/internal/k8s/api/v1"
	"github.com/110y/bootes/internal/k8s/store"
	"github.com/110y/bootes/internal/k8s/testutils"
)

var k8sClient client.Client

func TestMain(m *testing.M) {
	os.Exit(func() int {
		cli, done, err := testutils.TestK8SClient()
		if err != nil {
			fmt.Fprintf(os.Stdout, fmt.Sprintf("failed to create k8s client: %s", err))
			return 1
		}
		defer done()

		k8sClient = cli

		return m.Run()
	}())
}

func TestGetCluster(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	namespace := testutils.NewNamespace(t, ctx, k8sClient)

	fixtures := []*unstructured.Unstructured{
		{
			Object: map[string]interface{}{
				"kind":       api.ClusterKind,
				"apiVersion": api.GroupVersion.String(),
				"metadata": map[string]interface{}{
					"name":      "test-cluster-1",
					"namespace": namespace,
				},
				"spec": map[string]interface{}{
					"workloadSelector": map[string]interface{}{
						"labels": map[string]interface{}{
							"label-1": "1",
							"label-2": "2",
						},
					},
					"config": map[string]interface{}{
						"name":            "cluster-1",
						"connect_timeout": "1s",
						"type":            "LOGICAL_DNS",
						"load_assignment": map[string]interface{}{
							"cluster_name": "cluster-1",
							"endpoints": []map[string]interface{}{
								{
									"lb_endpoints": []map[string]interface{}{
										{
											"endpoint": map[string]interface{}{
												"address": map[string]interface{}{
													"socket_address": map[string]interface{}{
														"address":    "test-1.test.svc.cluster.local",
														"port_value": "10000",
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
		},
		{
			Object: map[string]interface{}{
				"kind":       api.ClusterKind,
				"apiVersion": api.GroupVersion.String(),
				"metadata": map[string]interface{}{
					"name":      "test-cluster-2",
					"namespace": namespace,
				},
				"spec": map[string]interface{}{
					"config": map[string]interface{}{
						"name":            "cluster-2",
						"connect_timeout": "1s",
						"type":            "LOGICAL_DNS",
						"load_assignment": map[string]interface{}{
							"cluster_name": "cluster-2",
							"endpoints": []map[string]interface{}{
								{
									"lb_endpoints": []map[string]interface{}{
										{
											"endpoint": map[string]interface{}{
												"address": map[string]interface{}{
													"socket_address": map[string]interface{}{
														"address":    "test-2.test.svc.cluster.local",
														"port_value": "10000",
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
		},
	}

	for _, f := range fixtures {
		if err := k8sClient.Create(ctx, f); err != nil {
			t.Fatalf("failed to create fixture: %s", err)
		}
	}

	s := store.New(k8sClient, k8sClient)

	tests := map[string]struct {
		expected *api.Cluster
		name     string
	}{
		"should get cluster": {
			name: "test-cluster-1",
			expected: &api.Cluster{
				Spec: api.ClusterSpec{
					WorkloadSelector: &api.WorkloadSelector{
						Labels: map[string]string{
							"label-1": "1",
							"label-2": "2",
						},
					},
					Config: &cluster.Cluster{
						Name:           "cluster-1",
						ConnectTimeout: &duration.Duration{Seconds: 1},
						ClusterDiscoveryType: &cluster.Cluster_Type{
							Type: cluster.Cluster_LOGICAL_DNS,
						},
						LoadAssignment: &endpoint.ClusterLoadAssignment{
							ClusterName: "cluster-1",
							Endpoints: []*endpoint.LocalityLbEndpoints{
								{
									LbEndpoints: []*endpoint.LbEndpoint{
										{
											HostIdentifier: &endpoint.LbEndpoint_Endpoint{
												Endpoint: &endpoint.Endpoint{
													Address: &core.Address{
														Address: &core.Address_SocketAddress{
															SocketAddress: &core.SocketAddress{
																Address: "test-1.test.svc.cluster.local",
																PortSpecifier: &core.SocketAddress_PortValue{
																	PortValue: 10000,
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
					},
				},
			},
		},
		"should get cluster even though workloadSelector is empty": {
			name: "test-cluster-2",
			expected: &api.Cluster{
				Spec: api.ClusterSpec{
					Config: &cluster.Cluster{
						Name:           "cluster-2",
						ConnectTimeout: &duration.Duration{Seconds: 1},
						ClusterDiscoveryType: &cluster.Cluster_Type{
							Type: cluster.Cluster_LOGICAL_DNS,
						},
						LoadAssignment: &endpoint.ClusterLoadAssignment{
							ClusterName: "cluster-2",
							Endpoints: []*endpoint.LocalityLbEndpoints{
								{
									LbEndpoints: []*endpoint.LbEndpoint{
										{
											HostIdentifier: &endpoint.LbEndpoint_Endpoint{
												Endpoint: &endpoint.Endpoint{
													Address: &core.Address{
														Address: &core.Address_SocketAddress{
															SocketAddress: &core.SocketAddress{
																Address: "test-2.test.svc.cluster.local",
																PortSpecifier: &core.SocketAddress_PortValue{
																	PortValue: 10000,
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
					},
				},
			},
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, err := s.GetCluster(ctx, test.name, namespace)
			if err != nil {
				t.Fatalf("error: %s", err)
			}

			if diff := cmp.Diff(test.expected, actual, testutils.CmpOptProtoTransformer); diff != "" {
				t.Errorf("\n(-expected, +actual)\n%s", diff)
			}
		})
	}
}

func TestListClustersByNamespace(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	namespace := testutils.NewNamespace(t, ctx, k8sClient)

	fixtures := []*unstructured.Unstructured{
		{
			Object: map[string]interface{}{
				"kind":       api.ClusterKind,
				"apiVersion": api.GroupVersion.String(),
				"metadata": map[string]interface{}{
					"name":      "test-cluster-1",
					"namespace": namespace,
				},
				"spec": map[string]interface{}{
					"workloadSelector": map[string]interface{}{
						"labels": map[string]interface{}{
							"label-1": "1",
							"label-2": "2",
						},
					},
					"config": map[string]interface{}{
						"name":            "cluster-1",
						"connect_timeout": "1s",
						"type":            "LOGICAL_DNS",
						"load_assignment": map[string]interface{}{
							"cluster_name": "cluster-1",
							"endpoints": []map[string]interface{}{
								{
									"lb_endpoints": []map[string]interface{}{
										{
											"endpoint": map[string]interface{}{
												"address": map[string]interface{}{
													"socket_address": map[string]interface{}{
														"address":    "test-1.test.svc.cluster.local",
														"port_value": "10000",
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
		},
		{
			Object: map[string]interface{}{
				"kind":       api.ClusterKind,
				"apiVersion": api.GroupVersion.String(),
				"metadata": map[string]interface{}{
					"name":      "test-cluster-2",
					"namespace": namespace,
				},
				"spec": map[string]interface{}{
					"config": map[string]interface{}{
						"name":            "cluster-2",
						"connect_timeout": "1s",
						"type":            "LOGICAL_DNS",
						"load_assignment": map[string]interface{}{
							"cluster_name": "cluster-2",
							"endpoints": []map[string]interface{}{
								{
									"lb_endpoints": []map[string]interface{}{
										{
											"endpoint": map[string]interface{}{
												"address": map[string]interface{}{
													"socket_address": map[string]interface{}{
														"address":    "test-2.test.svc.cluster.local",
														"port_value": "10000",
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
		},
	}

	for _, f := range fixtures {
		if err := k8sClient.Create(ctx, f); err != nil {
			t.Fatalf("failed to create fixture: %s", err)
		}
	}

	s := store.New(k8sClient, k8sClient)

	tests := map[string]struct {
		expected *api.ClusterList
	}{
		"should list clusters": {
			expected: &api.ClusterList{
				Items: []*api.Cluster{
					{
						Spec: api.ClusterSpec{
							WorkloadSelector: &api.WorkloadSelector{
								Labels: map[string]string{
									"label-1": "1",
									"label-2": "2",
								},
							},
							Config: &cluster.Cluster{
								Name:           "cluster-1",
								ConnectTimeout: &duration.Duration{Seconds: 1},
								ClusterDiscoveryType: &cluster.Cluster_Type{
									Type: cluster.Cluster_LOGICAL_DNS,
								},
								LoadAssignment: &endpoint.ClusterLoadAssignment{
									ClusterName: "cluster-1",
									Endpoints: []*endpoint.LocalityLbEndpoints{
										{
											LbEndpoints: []*endpoint.LbEndpoint{
												{
													HostIdentifier: &endpoint.LbEndpoint_Endpoint{
														Endpoint: &endpoint.Endpoint{
															Address: &core.Address{
																Address: &core.Address_SocketAddress{
																	SocketAddress: &core.SocketAddress{
																		Address: "test-1.test.svc.cluster.local",
																		PortSpecifier: &core.SocketAddress_PortValue{
																			PortValue: 10000,
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
							},
						},
					},
					{
						Spec: api.ClusterSpec{
							Config: &cluster.Cluster{
								Name:           "cluster-2",
								ConnectTimeout: &duration.Duration{Seconds: 1},
								ClusterDiscoveryType: &cluster.Cluster_Type{
									Type: cluster.Cluster_LOGICAL_DNS,
								},
								LoadAssignment: &endpoint.ClusterLoadAssignment{
									ClusterName: "cluster-2",
									Endpoints: []*endpoint.LocalityLbEndpoints{
										{
											LbEndpoints: []*endpoint.LbEndpoint{
												{
													HostIdentifier: &endpoint.LbEndpoint_Endpoint{
														Endpoint: &endpoint.Endpoint{
															Address: &core.Address{
																Address: &core.Address_SocketAddress{
																	SocketAddress: &core.SocketAddress{
																		Address: "test-2.test.svc.cluster.local",
																		PortSpecifier: &core.SocketAddress_PortValue{
																			PortValue: 10000,
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
							},
						},
					},
				},
			},
		},
	}

	for name, test := range tests {
		test := test

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, err := s.ListClustersByNamespace(ctx, namespace)
			if err != nil {
				t.Fatalf("failed %s", err)
			}

			if len(test.expected.Items) != len(actual.Items) {
				t.Fatal("Different number of Items found")
			}

			for i, a := range actual.Items {
				if diff := cmp.Diff(test.expected.Items[i], a, testutils.CmpOptProtoTransformer); diff != "" {
					t.Errorf("\n(-expected, +actual)\n%s", diff)
				}
			}
		})
	}
}

func TestGetListener(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	namespace := testutils.NewNamespace(t, ctx, k8sClient)

	fixtures := []*unstructured.Unstructured{
		{
			Object: map[string]interface{}{
				"kind":       api.ListenerKind,
				"apiVersion": api.GroupVersion.String(),
				"metadata": map[string]interface{}{
					"name":      "test-listener-1",
					"namespace": namespace,
				},
				"spec": map[string]interface{}{
					"workloadSelector": map[string]interface{}{
						"labels": map[string]interface{}{
							"label-1": "1",
							"label-2": "2",
						},
					},
					"config": map[string]interface{}{
						"address": map[string]interface{}{
							"socket_address": map[string]interface{}{
								"protocol":   "TCP",
								"address":    "0.0.0.0",
								"port_value": "10000",
							},
						},
						"filter_chains": []map[string]interface{}{
							{
								"filters": []map[string]interface{}{
									{
										"name": "envoy.http_connection_manager",
										"typed_config": map[string]interface{}{
											"@type":       "type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager",
											"stat_prefix": "ingress_http",
											"route_config": map[string]interface{}{
												"name": "route",
												"virtual_hosts": []map[string]interface{}{
													{
														"name":    "service",
														"domains": []string{"*"},
														"routes": []map[string]interface{}{
															{
																"match": map[string]interface{}{
																	"prefix": "/",
																},
																"route": map[string]interface{}{
																	"cluster": "upstream",
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
					},
				},
			},
		},
		{
			Object: map[string]interface{}{
				"kind":       api.ListenerKind,
				"apiVersion": api.GroupVersion.String(),
				"metadata": map[string]interface{}{
					"name":      "test-listener-2",
					"namespace": namespace,
				},
				"spec": map[string]interface{}{
					"config": map[string]interface{}{
						"address": map[string]interface{}{
							"socket_address": map[string]interface{}{
								"protocol":   "TCP",
								"address":    "0.0.0.0",
								"port_value": "10000",
							},
						},
						"filter_chains": []map[string]interface{}{
							{
								"filters": []map[string]interface{}{
									{
										"name": "envoy.http_connection_manager",
										"typed_config": map[string]interface{}{
											"@type":       "type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager",
											"stat_prefix": "ingress_http",
											"route_config": map[string]interface{}{
												"name": "route",
												"virtual_hosts": []map[string]interface{}{
													{
														"name":    "service",
														"domains": []string{"*"},
														"routes": []map[string]interface{}{
															{
																"match": map[string]interface{}{
																	"prefix": "/",
																},
																"route": map[string]interface{}{
																	"cluster": "upstream",
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
					},
				},
			},
		},
	}

	for _, f := range fixtures {
		if err := k8sClient.Create(ctx, f); err != nil {
			t.Fatalf("failed to create fixture: %s", err)
		}
	}

	s := store.New(k8sClient, k8sClient)

	cm, err := proto.Marshal(&hcm.HttpConnectionManager{
		StatPrefix: "ingress_http",
		RouteSpecifier: &hcm.HttpConnectionManager_RouteConfig{
			RouteConfig: &route.RouteConfiguration{
				Name: "route",
				VirtualHosts: []*route.VirtualHost{
					{
						Name:    "service",
						Domains: []string{"*"},
						Routes: []*route.Route{
							{
								Match: &route.RouteMatch{
									PathSpecifier: &route.RouteMatch_Prefix{
										Prefix: "/",
									},
								},
								Action: &route.Route_Route{
									Route: &route.RouteAction{
										ClusterSpecifier: &route.RouteAction_Cluster{
											Cluster: "upstream",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to marshal fixture proto: %s", err)
	}

	tests := map[string]struct {
		expected *api.Listener
		name     string
	}{
		"should get listener": {
			name: "test-listener-1",
			expected: &api.Listener{
				Spec: api.ListenerSpec{
					WorkloadSelector: &api.WorkloadSelector{
						Labels: map[string]string{
							"label-1": "1",
							"label-2": "2",
						},
					},
					Config: &listener.Listener{
						Address: &core.Address{
							Address: &core.Address_SocketAddress{
								SocketAddress: &core.SocketAddress{
									Protocol: core.SocketAddress_TCP,
									Address:  "0.0.0.0",
									PortSpecifier: &core.SocketAddress_PortValue{
										PortValue: 10000,
									},
								},
							},
						},
						FilterChains: []*listener.FilterChain{
							{
								Filters: []*listener.Filter{
									{
										Name: "envoy.http_connection_manager",
										ConfigType: &listener.Filter_TypedConfig{
											TypedConfig: &any.Any{
												TypeUrl: "type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager",
												Value:   cm,
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
		"should get cluster even though workloadSelector is empty": {
			name: "test-listener-2",
			expected: &api.Listener{
				Spec: api.ListenerSpec{
					Config: &listener.Listener{
						Address: &core.Address{
							Address: &core.Address_SocketAddress{
								SocketAddress: &core.SocketAddress{
									Protocol: core.SocketAddress_TCP,
									Address:  "0.0.0.0",
									PortSpecifier: &core.SocketAddress_PortValue{
										PortValue: 10000,
									},
								},
							},
						},
						FilterChains: []*listener.FilterChain{
							{
								Filters: []*listener.Filter{
									{
										Name: "envoy.http_connection_manager",
										ConfigType: &listener.Filter_TypedConfig{
											TypedConfig: &any.Any{
												TypeUrl: "type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager",
												Value:   cm,
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

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, err := s.GetListener(ctx, test.name, namespace)
			if err != nil {
				t.Fatalf("error: %s", err)
			}

			if diff := cmp.Diff(test.expected, actual, testutils.CmpOptProtoTransformer); diff != "" {
				t.Errorf("\n(-expected, +actual)\n%s", diff)
			}
		})
	}
}

func TestListListenersByNamespace(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	namespace := testutils.NewNamespace(t, ctx, k8sClient)

	fixtures := []*unstructured.Unstructured{
		{
			Object: map[string]interface{}{
				"kind":       api.ListenerKind,
				"apiVersion": api.GroupVersion.String(),
				"metadata": map[string]interface{}{
					"name":      "test-listener-1",
					"namespace": namespace,
				},
				"spec": map[string]interface{}{
					"workloadSelector": map[string]interface{}{
						"labels": map[string]interface{}{
							"label-1": "1",
							"label-2": "2",
						},
					},
					"config": map[string]interface{}{
						"address": map[string]interface{}{
							"socket_address": map[string]interface{}{
								"protocol":   "TCP",
								"address":    "0.0.0.0",
								"port_value": "10000",
							},
						},
						"filter_chains": []map[string]interface{}{
							{
								"filters": []map[string]interface{}{
									{
										"name": "envoy.http_connection_manager",
										"typed_config": map[string]interface{}{
											"@type":       "type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager",
											"stat_prefix": "ingress_http",
											"route_config": map[string]interface{}{
												"name": "route",
												"virtual_hosts": []map[string]interface{}{
													{
														"name":    "service",
														"domains": []string{"*"},
														"routes": []map[string]interface{}{
															{
																"match": map[string]interface{}{
																	"prefix": "/",
																},
																"route": map[string]interface{}{
																	"cluster": "upstream",
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
					},
				},
			},
		},
		{
			Object: map[string]interface{}{
				"kind":       api.ListenerKind,
				"apiVersion": api.GroupVersion.String(),
				"metadata": map[string]interface{}{
					"name":      "test-listener-2",
					"namespace": namespace,
				},
				"spec": map[string]interface{}{
					"config": map[string]interface{}{
						"address": map[string]interface{}{
							"socket_address": map[string]interface{}{
								"protocol":   "TCP",
								"address":    "0.0.0.0",
								"port_value": "10000",
							},
						},
						"filter_chains": []map[string]interface{}{
							{
								"filters": []map[string]interface{}{
									{
										"name": "envoy.http_connection_manager",
										"typed_config": map[string]interface{}{
											"@type":       "type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager",
											"stat_prefix": "ingress_http",
											"route_config": map[string]interface{}{
												"name": "route",
												"virtual_hosts": []map[string]interface{}{
													{
														"name":    "service",
														"domains": []string{"*"},
														"routes": []map[string]interface{}{
															{
																"match": map[string]interface{}{
																	"prefix": "/",
																},
																"route": map[string]interface{}{
																	"cluster": "upstream",
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
					},
				},
			},
		},
	}

	for _, f := range fixtures {
		if err := k8sClient.Create(ctx, f); err != nil {
			t.Fatalf("failed to create fixture: %s", err)
		}
	}

	s := store.New(k8sClient, k8sClient)

	cm, err := proto.Marshal(&hcm.HttpConnectionManager{
		StatPrefix: "ingress_http",
		RouteSpecifier: &hcm.HttpConnectionManager_RouteConfig{
			RouteConfig: &route.RouteConfiguration{
				Name: "route",
				VirtualHosts: []*route.VirtualHost{
					{
						Name:    "service",
						Domains: []string{"*"},
						Routes: []*route.Route{
							{
								Match: &route.RouteMatch{
									PathSpecifier: &route.RouteMatch_Prefix{
										Prefix: "/",
									},
								},
								Action: &route.Route_Route{
									Route: &route.RouteAction{
										ClusterSpecifier: &route.RouteAction_Cluster{
											Cluster: "upstream",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to marshal fixture proto: %s", err)
	}

	tests := map[string]struct {
		expected *api.ListenerList
	}{
		"should list listeners": {
			&api.ListenerList{
				Items: []*api.Listener{
					{
						Spec: api.ListenerSpec{
							WorkloadSelector: &api.WorkloadSelector{
								Labels: map[string]string{
									"label-1": "1",
									"label-2": "2",
								},
							},
							Config: &listener.Listener{
								Address: &core.Address{
									Address: &core.Address_SocketAddress{
										SocketAddress: &core.SocketAddress{
											Protocol: core.SocketAddress_TCP,
											Address:  "0.0.0.0",
											PortSpecifier: &core.SocketAddress_PortValue{
												PortValue: 10000,
											},
										},
									},
								},
								FilterChains: []*listener.FilterChain{
									{
										Filters: []*listener.Filter{
											{
												Name: "envoy.http_connection_manager",
												ConfigType: &listener.Filter_TypedConfig{
													TypedConfig: &any.Any{
														TypeUrl: "type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager",
														Value:   cm,
													},
												},
											},
										},
									},
								},
							},
						},
					},
					{
						Spec: api.ListenerSpec{
							Config: &listener.Listener{
								Address: &core.Address{
									Address: &core.Address_SocketAddress{
										SocketAddress: &core.SocketAddress{
											Protocol: core.SocketAddress_TCP,
											Address:  "0.0.0.0",
											PortSpecifier: &core.SocketAddress_PortValue{
												PortValue: 10000,
											},
										},
									},
								},
								FilterChains: []*listener.FilterChain{
									{
										Filters: []*listener.Filter{
											{
												Name: "envoy.http_connection_manager",
												ConfigType: &listener.Filter_TypedConfig{
													TypedConfig: &any.Any{
														TypeUrl: "type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager",
														Value:   cm,
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
		},
	}

	for name, test := range tests {
		test := test

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, err := s.ListListenersByNamespace(ctx, namespace)
			if err != nil {
				t.Fatalf("failed %s", err)
			}

			if len(test.expected.Items) != len(actual.Items) {
				t.Fatal("Different number of Items found")
			}

			for i, a := range actual.Items {
				if diff := cmp.Diff(test.expected.Items[i], a, testutils.CmpOptProtoTransformer); diff != "" {
					t.Errorf("\n(-expected, +actual)\n%s", diff)
				}
			}
		})
	}
}

func TestGetRoute(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	namespace := testutils.NewNamespace(t, ctx, k8sClient)

	fixtures := []*unstructured.Unstructured{
		{
			Object: map[string]interface{}{
				"kind":       api.RouteKind,
				"apiVersion": api.GroupVersion.String(),
				"metadata": map[string]interface{}{
					"name":      "test-route-1",
					"namespace": namespace,
				},
				"spec": map[string]interface{}{
					"workloadSelector": map[string]interface{}{
						"labels": map[string]interface{}{
							"label-1": "1",
							"label-2": "2",
						},
					},
					"config": map[string]interface{}{
						"name": "route",
						"virtual_hosts": []map[string]interface{}{
							{
								"name":    "service",
								"domains": []string{"*"},
								"routes": []map[string]interface{}{
									{

										"name": "cluster-1",
										"route": map[string]interface{}{
											"cluster": "cluster-1",
										},
										"match": map[string]interface{}{
											"prefix": "/",
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

	for _, f := range fixtures {
		if err := k8sClient.Create(ctx, f); err != nil {
			t.Fatalf("failed to create fixture: %s", err)
		}
	}

	s := store.New(k8sClient, k8sClient)

	tests := map[string]struct {
		expected *api.Route
		name     string
	}{
		"should get route": {
			name: "test-route-1",
			expected: &api.Route{
				Spec: api.RouteSpec{
					WorkloadSelector: &api.WorkloadSelector{
						Labels: map[string]string{
							"label-1": "1",
							"label-2": "2",
						},
					},
					Config: &route.RouteConfiguration{
						Name: "route",
						VirtualHosts: []*route.VirtualHost{
							{
								Name:    "service",
								Domains: []string{"*"},
								Routes: []*route.Route{
									{
										Name: "cluster-1",
										Match: &route.RouteMatch{
											PathSpecifier: &route.RouteMatch_Prefix{
												Prefix: "/",
											},
										},
										Action: &route.Route_Route{
											Route: &route.RouteAction{
												ClusterSpecifier: &route.RouteAction_Cluster{
													Cluster: "cluster-1",
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

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			actual, err := s.GetRoute(ctx, test.name, namespace)
			if err != nil {
				t.Fatalf("error: %s", err)
			}

			if diff := cmp.Diff(test.expected, actual, testutils.CmpOptProtoTransformer); diff != "" {
				t.Errorf("\n(-expected, +actual)\n%s", diff)
			}
		})
	}
}

func TestListRoutesByNamespace(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	namespace := testutils.NewNamespace(t, ctx, k8sClient)

	fixtures := []*unstructured.Unstructured{
		{
			Object: map[string]interface{}{
				"kind":       api.RouteKind,
				"apiVersion": api.GroupVersion.String(),
				"metadata": map[string]interface{}{
					"name":      "test-route-1",
					"namespace": namespace,
				},
				"spec": map[string]interface{}{
					"workloadSelector": map[string]interface{}{
						"labels": map[string]interface{}{
							"label-1": "1",
							"label-2": "2",
						},
					},
					"config": map[string]interface{}{
						"name": "route",
						"virtual_hosts": []map[string]interface{}{
							{
								"name":    "service",
								"domains": []string{"*"},
								"routes": []map[string]interface{}{
									{

										"name": "cluster-1",
										"route": map[string]interface{}{
											"cluster": "cluster-1",
										},
										"match": map[string]interface{}{
											"prefix": "/",
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

	for _, f := range fixtures {
		if err := k8sClient.Create(ctx, f); err != nil {
			t.Fatalf("failed to create fixture: %s", err)
		}
	}

	s := store.New(k8sClient, k8sClient)

	tests := map[string]struct {
		expected *api.RouteList
	}{
		"should list clusters": {
			expected: &api.RouteList{
				Items: []*api.Route{
					{
						Spec: api.RouteSpec{
							WorkloadSelector: &api.WorkloadSelector{
								Labels: map[string]string{
									"label-1": "1",
									"label-2": "2",
								},
							},
							Config: &route.RouteConfiguration{
								Name: "route",
								VirtualHosts: []*route.VirtualHost{
									{
										Name:    "service",
										Domains: []string{"*"},
										Routes: []*route.Route{
											{
												Name: "cluster-1",
												Match: &route.RouteMatch{
													PathSpecifier: &route.RouteMatch_Prefix{
														Prefix: "/",
													},
												},
												Action: &route.Route_Route{
													Route: &route.RouteAction{
														ClusterSpecifier: &route.RouteAction_Cluster{
															Cluster: "cluster-1",
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
			},
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, err := s.ListRoutesByNamespace(ctx, namespace)
			if err != nil {
				t.Fatalf("failed %s", err)
			}

			if len(test.expected.Items) != len(actual.Items) {
				t.Fatal("Different number of Items found")
			}

			for i, a := range actual.Items {
				if diff := cmp.Diff(test.expected.Items[i], a, testutils.CmpOptProtoTransformer); diff != "" {
					t.Errorf("\n(-expected, +actual)\n%s", diff)
				}
			}
		})
	}
}

func TestListPodsByNamespace(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	namespace := testutils.NewNamespace(t, ctx, k8sClient)

	pod1 := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-1",
			Namespace: namespace,
			Labels: map[string]string{
				"app":  "envoy",
				"test": "1",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "envoy",
					Image: "envoyproxy/envoy:latest",
				},
			},
		},
	}

	pod2 := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-2",
			Namespace: namespace,
			Labels: map[string]string{
				"app":  "envoy",
				"test": "2",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "envoy",
					Image: "envoyproxy/envoy:latest",
				},
			},
		},
	}

	fixtures := []corev1.Pod{
		pod1,
		pod2,
	}

	for _, f := range fixtures {
		if err := k8sClient.Create(ctx, &f); err != nil {
			t.Fatalf("failed to create fixture: %s", err)
		}
	}

	tests := map[string]struct {
		expected *corev1.PodList
		options  []store.ListOption
	}{
		"should list all pods": {
			expected: &corev1.PodList{
				Items: fixtures,
			},
		},
		"should list pod1": {
			expected: &corev1.PodList{
				Items: []corev1.Pod{pod1},
			},
			options: []store.ListOption{
				store.WithLabelFilter(map[string]string{
					"app":  "envoy",
					"test": "1",
				}),
			},
		},
		"should list all pods with label selectors": {
			expected: &corev1.PodList{
				Items: fixtures,
			},
			options: []store.ListOption{
				store.WithLabelFilter(map[string]string{
					"app": "envoy",
				}),
			},
		},
	}

	s := store.New(k8sClient, k8sClient)

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, err := s.ListPodsByNamespace(ctx, namespace, test.options...)
			if err != nil {
				t.Fatalf("failed %s", err)
			}

			if diff := cmp.Diff(test.expected.Items, actual.Items, testutils.CmpOptPodComparer); diff != "" {
				t.Errorf("\n(-expected, +actual)\n%s", diff)
			}
		})
	}
}
