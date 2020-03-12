package store

import (
	"context"

	api "github.com/110y/bootes/internal/k8s/api/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Store interface {
	GetCluster(ctx context.Context, name, namespace string) (*api.Cluster, error)
	ListPodsByNamespace(ctx context.Context, namespace string) (corev1.PodList, error)
	ListClustersByNamespace(ctx context.Context, namespace string) (api.ClusterList, error)
}

type store struct {
	client client.Client
	reader client.Reader
}

func New(c client.Client, reader client.Reader) Store {
	return &store{
		client: c,
		reader: reader,
	}
}

func (s *store) GetCluster(ctx context.Context, name, namespace string) (*api.Cluster, error) {
	key := client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}

	var cluster api.Cluster
	if err := s.client.Get(ctx, key, &cluster); err != nil {
		if apierrors.IsNotFound(err) {
		}

		// TODO:
		return nil, nil
	}

	return &cluster, nil
}

func (s *store) ListClustersByNamespace(ctx context.Context, namespace string) (api.ClusterList, error) {
	var clusters api.ClusterList
	err := s.client.List(ctx, &clusters, &client.ListOptions{
		Namespace: namespace,
	})
	if err != nil {
		// TODO:
		return clusters, err
	}

	return clusters, nil
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
