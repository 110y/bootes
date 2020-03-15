package store

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	api "github.com/110y/bootes/internal/k8s/api/v1"
	envoyapi "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/golang/protobuf/jsonpb"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Store interface {
	GetCluster(ctx context.Context, name, namespace string) (*api.Cluster, error)
	ListPodsByNamespace(ctx context.Context, namespace string) (corev1.PodList, error)
	ListClustersByNamespace(ctx context.Context, namespace string) (*api.ClusterList, error)
}

type store struct {
	client      client.Client
	reader      client.Reader
	unmarshaler *jsonpb.Unmarshaler
}

func New(c client.Client, reader client.Reader) Store {
	return &store{
		client: c,
		reader: reader,
		unmarshaler: &jsonpb.Unmarshaler{
			AllowUnknownFields: true,
		},
	}
}

func (s *store) GetCluster(ctx context.Context, name, namespace string) (*api.Cluster, error) {
	return nil, nil
}

func (s *store) ListClustersByNamespace(ctx context.Context, namespace string) (*api.ClusterList, error) {
	clusters := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"kind":       "Cluster",
			"apiVersion": api.GroupVersion.String(),
		},
	}
	err := s.client.List(ctx, clusters, &client.ListOptions{
		Namespace: namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters: %w", err)
	}

	items := make([]*api.Cluster, len(clusters.Items))
	for i, c := range clusters.Items {
		spec, ok := c.Object["spec"]
		if !ok {
			return nil, fmt.Errorf("spec not found")
		}

		sp, ok := spec.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid spec form")
		}

		config, ok := sp["config"]
		if !ok {
			return nil, fmt.Errorf("spec.config not found")
		}

		j, err := json.Marshal(config)
		if err != nil {
			return nil, fmt.Errorf("failed to parse spec.config: %s", err)
		}

		cl := &envoyapi.Cluster{}
		if err := s.unmarshaler.Unmarshal(bytes.NewBuffer(j), cl); err != nil {
			return nil, fmt.Errorf("failed to unmarshal spec.config: %s", err)
		}

		items[i] = &api.Cluster{
			Spec: api.ClusterSpec{
				Config: cl,
			},
		}
	}

	return &api.ClusterList{
		Items: items,
	}, nil
}

func (s *store) ListPodsByNamespace(ctx context.Context, namespace string) (corev1.PodList, error) {
	var pods corev1.PodList
	err := s.client.List(ctx, &pods, &client.ListOptions{
		Namespace: namespace,
	})
	if err != nil {
		// TODO:
		return pods, err
	}

	return pods, nil
}
