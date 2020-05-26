package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	envoyapi "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/encoding/protojson"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"

	api "github.com/110y/bootes/internal/k8s/api/v1"
	"github.com/110y/bootes/internal/observer/trace"
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

var _ Store = (*store)(nil)

type Store interface {
	GetCluster(ctx context.Context, name, namespace string) (*api.Cluster, error)
	ListClustersByNamespace(ctx context.Context, namespace string) (*api.ClusterList, error)
	GetListener(ctx context.Context, name, namespace string) (*api.Listener, error)
	ListListenersByNamespace(ctx context.Context, namespace string) (*api.ListenerList, error)
	GetRoute(ctx context.Context, name, namespace string) (*api.Route, error)
	ListRoutesByNamespace(ctx context.Context, namespace string) (*api.RouteList, error)
	GetEndpoint(ctx context.Context, name, namespace string) (*api.Endpoint, error)
	ListEndpointsByNamespace(ctx context.Context, namespace string) (*api.EndpointList, error)
	GetPod(ctx context.Context, name, namespace string) (*corev1.Pod, error)
	ListPodsByNamespace(ctx context.Context, namespace string, options ...ListOption) (*corev1.PodList, error)
}

type store struct {
	client      client.Client
	reader      client.Reader
	unmarshaler *protojson.UnmarshalOptions
}

func New(c client.Client, reader client.Reader) Store {
	return &store{
		client: c,
		reader: reader,
		unmarshaler: &protojson.UnmarshalOptions{
			AllowPartial:   false,
			DiscardUnknown: true,
		},
	}
}

func (s *store) GetCluster(ctx context.Context, name, namespace string) (*api.Cluster, error) {
	ctx, span := trace.NewSpan(ctx, "Store.GetCluster")
	defer span.End()

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
	ctx, span := trace.NewSpan(ctx, "Store.ListClustersByNamespace")
	defer span.End()

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

func (s *store) GetListener(ctx context.Context, name, namespace string) (*api.Listener, error) {
	ctx, span := trace.NewSpan(ctx, "Store.GetListener")
	defer span.End()

	key := client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}

	listener := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       api.ListenerKind,
			"apiVersion": api.GroupVersion.String(),
		},
	}

	if err := s.client.Get(ctx, key, listener); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("failed to get route: %w", err)
	}

	l, err := s.unmarshalListener(listener.Object)
	if err != nil {
		return nil, err
	}

	return l, nil
}

func (s *store) ListListenersByNamespace(ctx context.Context, namespace string) (*api.ListenerList, error) {
	ctx, span := trace.NewSpan(ctx, "Store.ListListenersByNamespace")
	defer span.End()

	listeners := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"kind":       api.ListenerKind,
			"apiVersion": api.GroupVersion.String(),
		},
	}
	err := s.client.List(ctx, listeners, &client.ListOptions{
		Namespace: namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list listeners: %w", err)
	}

	items := make([]*api.Listener, len(listeners.Items))
	for i, c := range listeners.Items {
		listener, err := s.unmarshalListener(c.Object)
		if err != nil {
			return nil, err
		}

		items[i] = listener
	}

	return &api.ListenerList{
		Items: items,
	}, nil
}

func (s *store) GetRoute(ctx context.Context, name, namespace string) (*api.Route, error) {
	ctx, span := trace.NewSpan(ctx, "Store.GetRoute")
	defer span.End()

	key := client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}

	route := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       api.RouteKind,
			"apiVersion": api.GroupVersion.String(),
		},
	}

	if err := s.client.Get(ctx, key, route); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("failed to get route: %w", err)
	}

	r, err := api.UnmarshalRouteObject(route.Object)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (s *store) ListRoutesByNamespace(ctx context.Context, namespace string) (*api.RouteList, error) {
	ctx, span := trace.NewSpan(ctx, "Store.ListRoutesByNamespace")
	defer span.End()

	routes := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"kind":       api.RouteKind,
			"apiVersion": api.GroupVersion.String(),
		},
	}
	err := s.client.List(ctx, routes, &client.ListOptions{
		Namespace: namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list routes: %w", err)
	}

	items := make([]*api.Route, len(routes.Items))
	for i, c := range routes.Items {
		route, err := api.UnmarshalRouteObject(c.Object)
		if err != nil {
			return nil, err
		}

		items[i] = route
	}

	return &api.RouteList{
		Items: items,
	}, nil
}

func (s *store) GetEndpoint(ctx context.Context, name, namespace string) (*api.Endpoint, error) {
	ctx, span := trace.NewSpan(ctx, "Store.GetEndpoint")
	defer span.End()

	key := client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}

	endpoint := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       api.EndpointKind,
			"apiVersion": api.GroupVersion.String(),
		},
	}

	if err := s.client.Get(ctx, key, endpoint); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("failed to get endpoint: %w", err)
	}

	e, err := s.unmarshalEndpoint(endpoint.Object)
	if err != nil {
		return nil, err
	}

	return e, nil
}

func (s *store) ListEndpointsByNamespace(ctx context.Context, namespace string) (*api.EndpointList, error) {
	ctx, span := trace.NewSpan(ctx, "Store.ListEndpointsByNamespace")
	defer span.End()

	routes := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"kind":       api.EndpointKind,
			"apiVersion": api.GroupVersion.String(),
		},
	}
	err := s.client.List(ctx, routes, &client.ListOptions{
		Namespace: namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list endpoints: %w", err)
	}

	items := make([]*api.Endpoint, len(routes.Items))
	for i, c := range routes.Items {
		endpoint, err := s.unmarshalEndpoint(c.Object)
		if err != nil {
			return nil, err
		}

		items[i] = endpoint
	}

	return &api.EndpointList{
		Items: items,
	}, nil
}

func (s *store) GetPod(ctx context.Context, name, namespace string) (*corev1.Pod, error) {
	ctx, span := trace.NewSpan(ctx, "Store.GetPod")
	defer span.End()

	key := client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}

	var pod corev1.Pod
	if err := s.reader.Get(ctx, key, &pod); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("failed to get pod: %w", err)
	}

	return &pod, nil
}

func (s *store) ListPodsByNamespace(ctx context.Context, namespace string, options ...ListOption) (*corev1.PodList, error) {
	ctx, span := trace.NewSpan(ctx, "Store.ListPodsByNamespace")
	defer span.End()

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
	config, err := unmarshalEnvoyConfig(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal envoy configuration: %w", err)
	}

	cluster := &envoyapi.Cluster{}
	if err := s.unmarshaler.Unmarshal(config, proto.MessageV2(cluster)); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spec.config: %w", err)
	}

	return cluster, nil
}

func (s *store) unmarshalListener(object map[string]interface{}) (*api.Listener, error) {
	spec, err := extractSpecFromObject(object)
	if err != nil {
		return nil, err
	}

	config, err := s.unmarshalListenerConfig(spec)
	if err != nil {
		return nil, err
	}

	selector, err := unmarshalWorkloadSelector(spec)
	if err != nil && !errors.Is(err, errWorkloadSelectorNotFound) {
		return nil, err
	}

	return &api.Listener{
		Spec: api.ListenerSpec{
			WorkloadSelector: selector,
			Config:           config,
		},
	}, nil
}

func (s *store) unmarshalListenerConfig(spec map[string]interface{}) (*envoyapi.Listener, error) {
	config, err := unmarshalEnvoyConfig(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal envoy configuration: %w", err)
	}

	listener := &envoyapi.Listener{}
	if err := s.unmarshaler.Unmarshal(config, proto.MessageV2(listener)); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spec.config: %w", err)
	}

	return listener, nil
}

func (s *store) unmarshalEndpoint(object map[string]interface{}) (*api.Endpoint, error) {
	spec, err := extractSpecFromObject(object)
	if err != nil {
		return nil, err
	}

	config, err := s.unmarshalEndpointConfig(spec)
	if err != nil {
		return nil, err
	}

	selector, err := unmarshalWorkloadSelector(spec)
	if err != nil && !errors.Is(err, errWorkloadSelectorNotFound) {
		return nil, err
	}

	return &api.Endpoint{
		Spec: api.EndpointSpec{
			WorkloadSelector: selector,
			Config:           config,
		},
	}, nil
}

func (s *store) unmarshalEndpointConfig(spec map[string]interface{}) (*envoyapi.ClusterLoadAssignment, error) {
	config, err := unmarshalEnvoyConfig(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal envoy configuration: %w", err)
	}

	endpoint := &envoyapi.ClusterLoadAssignment{}
	if err := s.unmarshaler.Unmarshal(config, proto.MessageV2(endpoint)); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spec.config: %w", err)
	}

	return endpoint, nil
}

func unmarshalEnvoyConfig(spec map[string]interface{}) ([]byte, error) {
	config, ok := spec["config"]
	if !ok {
		return nil, fmt.Errorf("spec.config not found")
	}

	j, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse spec.config: %w", err)
	}

	return j, nil
}
