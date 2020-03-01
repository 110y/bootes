package store

import (
	"context"

	api "github.com/110y/bootes/internal/k8s/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ Store = &store{}

type Store interface {
	ListPodsByNamespace(ctx context.Context, namespace string) (*api.ClusterList, error)
	GetCluster(ctx context.Context, name, namespace string) (*api.Cluster, error)
}

type store struct {
	client client.Client
}

func NewStore(c client.Client) *store {
	return &store{client: c}
}

func (s *store) GetCluster(ctx context.Context, name, namespace string) (*api.Cluster, error) {
	key := client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}

	var cluster api.Cluster
	if err := s.client.Get(ctx, key, &cluster); err != nil {
		err = client.IgnoreNotFound(err)
		if err != nil {
			// TODO:
			return nil, err
		}
		// TODO:
		return nil, nil
	}

	return &cluster, nil
}

func (s *store) ListPodsByNamespace(ctx context.Context, namespace string) (*api.ClusterList, error) {
	var clusterList api.ClusterList
	err := s.client.List(ctx, &clusterList, &client.ListOptions{
		Namespace: namespace,
	})
	if err != nil {
		// TODO:
		return nil, err
	}

	return &clusterList, nil
}
