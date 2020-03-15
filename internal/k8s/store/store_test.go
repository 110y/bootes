package store_test

import (
	"context"
	"testing"

	apiv1 "github.com/110y/bootes/internal/k8s/api/v1"
	"github.com/110y/bootes/internal/k8s/store"
	"github.com/110y/bootes/internal/k8s/testutils"
	envoyapi "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestGetCluster(t *testing.T) {
	tests := map[string]struct {
		cluster   *apiv1.Cluster
		name      string
		namespace string
	}{
		"": {
			name:      "foo",
			namespace: "bar",
		},
	}

	cli, teardown := testutils.SetupEnvtest(t)
	defer teardown()

	s := store.New(cli, cli)

	ctx := context.Background()
	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			_, err := s.GetCluster(ctx, test.name, test.namespace)
			if err != nil {
				t.Errorf("failed %s", err)
			}
		})
	}
}

func TestListClusters(t *testing.T) {
	tests := map[string]struct {
		expected *apiv1.ClusterList
	}{
		"should list clusters": {
			expected: &apiv1.ClusterList{
				Items: []*apiv1.Cluster{
					&apiv1.Cluster{
						Spec: apiv1.ClusterSpec{
							Config: &envoyapi.Cluster{
								Name:           "cluster-1",
								ConnectTimeout: &duration.Duration{Seconds: 1},
								ClusterDiscoveryType: &envoyapi.Cluster_Type{
									Type: envoyapi.Cluster_LOGICAL_DNS,
								},
								LoadAssignment: &envoyapi.ClusterLoadAssignment{
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
					&apiv1.Cluster{
						Spec: apiv1.ClusterSpec{
							Config: &envoyapi.Cluster{
								Name:           "cluster-2",
								ConnectTimeout: &duration.Duration{Seconds: 1},
								ClusterDiscoveryType: &envoyapi.Cluster_Type{
									Type: envoyapi.Cluster_LOGICAL_DNS,
								},
								LoadAssignment: &envoyapi.ClusterLoadAssignment{
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

	cli, teardown := testutils.SetupEnvtest(t)
	defer teardown()

	ctx := context.Background()

	fixtures := []*unstructured.Unstructured{
		&unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       "Cluster",
				"apiVersion": apiv1.GroupVersion.String(),
				"metadata": map[string]interface{}{
					"name":      "test-cluster-1",
					"namespace": "test",
				},
				"spec": map[string]interface{}{
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
		&unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       "Cluster",
				"apiVersion": apiv1.GroupVersion.String(),
				"metadata": map[string]interface{}{
					"name":      "test-cluster-2",
					"namespace": "test",
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
		if err := cli.Create(ctx, f); err != nil {
			t.Fatalf("failed to create fixture: %s", err)
		}
	}

	s := store.New(cli, cli)

	for name, test := range tests {
		test := test

		t.Run(name, func(t *testing.T) {
			actual, err := s.ListClustersByNamespace(ctx, "test")
			if err != nil {
				t.Errorf("failed %s", err)
			}

			if diff := cmp.Diff(test.expected, actual); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
}
