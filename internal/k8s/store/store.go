package store

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	api "github.com/110y/bootes/internal/k8s/api/v1"
	envoyapi "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/golang/protobuf/jsonpb"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	ErrNotFound = errors.New("resource not found")

	errWorkloadSelectorNotFound = errors.New("workloadSelector not found")
)

type ListOption func(*listOption)

type listOption struct {
	filterLabels map[string]string
}

func WithLabelFilter(labels map[string]string) ListOption {
	return func(opt *listOption) {
		opt.filterLabels = labels
	}
}

type Store interface {
	GetCluster(ctx context.Context, name, namespace string) (*api.Cluster, error)
	ListClustersByNamespace(ctx context.Context, namespace string) (*api.ClusterList, error)
	ListPodsByNamespace(ctx context.Context, namespace string, options ...ListOption) (*corev1.PodList, error)
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
	key := client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}

	cluster := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       api.ClusterKind,
			"apiVersion": api.GroupVersion.String(),
		},
	}

	if err := s.client.Get(ctx, key, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("failed to get cluster: %w", err)
	}

	c, err := s.unmarshalCluster(cluster.Object)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (s *store) ListClustersByNamespace(ctx context.Context, namespace string) (*api.ClusterList, error) {
	clusters := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"kind":       api.ClusterKind,
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
		cluster, err := s.unmarshalCluster(c.Object)
		if err != nil {
			return nil, err
		}

		items[i] = cluster
	}

	return &api.ClusterList{
		Items: items,
	}, nil
}

func (s *store) ListPodsByNamespace(ctx context.Context, namespace string, options ...ListOption) (*corev1.PodList, error) {
	opt := &listOption{}
	for _, o := range options {
		o(opt)
	}

	lo := &client.ListOptions{
		Namespace: namespace,
	}

	lenLabels := len(opt.filterLabels)
	if lenLabels != 0 {
		requirements := make([]labels.Requirement, lenLabels)

		i := 0
		for key, val := range opt.filterLabels {
			r, err := labels.NewRequirement(key, selection.Equals, []string{val})
			if err != nil {
				return nil, fmt.Errorf("failed to use labels.Requirement")
			}

			requirements[i] = *r
			i++
		}

		lo.LabelSelector = labels.Everything().Add(requirements...)
	}

	var pods corev1.PodList
	err := s.client.List(ctx, &pods, lo)
	if err != nil {
		return nil, err
	}

	return &pods, nil
}

func extractSpecFromObject(object map[string]interface{}) (map[string]interface{}, error) {
	spec, ok := object["spec"]
	if !ok {
		return nil, fmt.Errorf("spec not found")
	}

	s, ok := spec.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid spec form")
	}

	return s, nil
}

func unmarshalWorkloadSelector(spec map[string]interface{}) (*api.WorkloadSelector, error) {
	selector, ok := spec["workloadSelector"]
	if !ok {
		return nil, errWorkloadSelectorNotFound
	}

	j, err := json.Marshal(selector)
	if err != nil {
		return nil, fmt.Errorf("failed to parse spec.workloadSelector: %w", err)
	}

	var ws api.WorkloadSelector
	if err := json.Unmarshal(j, &ws); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spec.workloadSelector: %w", err)
	}

	return &ws, nil
}

func (s *store) unmarshalCluster(object map[string]interface{}) (*api.Cluster, error) {
	spec, err := extractSpecFromObject(object)
	if err != nil {
		return nil, err
	}

	config, err := s.unmarshalClusterConfig(spec)
	if err != nil {
		return nil, err
	}

	selector, err := unmarshalWorkloadSelector(spec)
	if err != nil && !errors.Is(err, errWorkloadSelectorNotFound) {
		return nil, err
	}

	return &api.Cluster{
		Spec: api.ClusterSpec{
			WorkloadSelector: selector,
			Config:           config,
		},
	}, nil
}

func (s *store) unmarshalClusterConfig(spec map[string]interface{}) (*envoyapi.Cluster, error) {
	config, ok := spec["config"]
	if !ok {
		return nil, fmt.Errorf("spec.config not found")
	}

	j, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse spec.config: %w", err)
	}

	cl := &envoyapi.Cluster{}
	if err := s.unmarshaler.Unmarshal(bytes.NewBuffer(j), cl); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spec.config: %w", err)
	}

	return cl, nil
}