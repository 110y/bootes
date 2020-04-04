package xds

import (
	"context"
	"fmt"

	api "github.com/110y/bootes/internal/k8s/api/v1"
	"github.com/110y/bootes/internal/k8s/store"
	"github.com/110y/bootes/internal/observer/trace"
	"github.com/110y/bootes/internal/xds/cache"
	envoyapi "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/pkg/server"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
)

var _ server.Callbacks = (*callbacks)(nil)

type callbacks struct {
	cache                  cache.Cache
	store                  store.Store
	loggerOnStreamOpen     logr.Logger
	loggerOnStreamClosed   logr.Logger
	loggerOnStreamRequest  logr.Logger
	loggerOnStreamResponse logr.Logger
	loggerOnFetchRequest   logr.Logger
	loggerOnFetchResponse  logr.Logger
}

func newCallbacks(c cache.Cache, s store.Store, l logr.Logger) *callbacks {
	return &callbacks{
		cache:                  c,
		store:                  s,
		loggerOnStreamOpen:     l.WithName("on_stream_open"),
		loggerOnStreamClosed:   l.WithName("on_stream_closed"),
		loggerOnStreamRequest:  l.WithName("on_stream_request"),
		loggerOnStreamResponse: l.WithName("on_stream_response"),
		loggerOnFetchRequest:   l.WithName("on_fetch_request"),
		loggerOnFetchResponse:  l.WithName("on_fetch_response"),
	}
}

func (c *callbacks) OnStreamOpen(_ context.Context, streamID int64, _ string) error {
	streamLogger(c.loggerOnStreamOpen, streamID).Info("open")
	return nil
}

func (c *callbacks) OnStreamClosed(streamID int64) {
	streamLogger(c.loggerOnStreamClosed, streamID).Info("closed")
}

func (c *callbacks) OnStreamRequest(streamID int64, req *envoyapi.DiscoveryRequest) error {
	ctx, span := trace.NewSpan(context.Background(), "Callbacks.OnStreamRequest")
	defer span.End()

	version := uuid.New().String()
	logger := requestLogger(streamLogger(c.loggerOnStreamRequest, streamID), req).WithValues("version", version)

	node := req.GetNode().GetId()
	if node == "" {
		logger.Info("empty node id passed")
		return fmt.Errorf("empty node id")
	}

	if c.cache.IsCachedNode(node) {
		// NOTE: use cache, no need to fetch resources again.
		return nil
	}

	name, namespace := store.ToNamespacedName(node)

	pod, err := c.store.GetPod(ctx, name, namespace)
	if err != nil {
		logger.Info("pod not found by node id")
		return fmt.Errorf("pod not found by node id")
	}

	clusters, err := c.listClustersByNodeAndLabels(ctx, namespace, pod.Labels)
	if err != nil {
		msg := "failed to list cluster configurations"
		logger.Error(err, msg)
		return fmt.Errorf("%s: %w", msg, err)
	}

	listeners, err := c.listListenersByNodeAndLabels(ctx, namespace, pod.Labels)
	if err != nil {
		msg := "failed to list listener configurations"
		logger.Error(err, msg)
		return fmt.Errorf("%s: %w", msg, err)
	}

	if err := c.cache.UpdateAllResources(ctx, node, version, clusters, listeners); err != nil {
		msg := "failed to update resources"
		logger.Error(err, msg)
		return fmt.Errorf("%s: %w", msg, err)
	}

	return nil
}

func (c *callbacks) OnStreamResponse(streamID int64, req *envoyapi.DiscoveryRequest, _ *envoyapi.DiscoveryResponse) {
	streamRequestLog(c.loggerOnStreamResponse, streamID, req)
}

func (c callbacks) OnFetchRequest(_ context.Context, req *envoyapi.DiscoveryRequest) error {
	requestLog(c.loggerOnFetchRequest, req)
	return nil
}

func (c *callbacks) OnFetchResponse(req *envoyapi.DiscoveryRequest, _ *envoyapi.DiscoveryResponse) {
	requestLog(c.loggerOnFetchResponse, req)
}

func (c *callbacks) listClustersByNodeAndLabels(ctx context.Context, namespace string, labels map[string]string) ([]*api.Cluster, error) {
	clusters, err := c.store.ListClustersByNamespace(ctx, namespace)
	if err != nil {
		return nil, err
	}

	return store.FilterClustersByLabels(clusters.Items, labels), nil
}

func (c *callbacks) listListenersByNodeAndLabels(ctx context.Context, namespace string, labels map[string]string) ([]*api.Listener, error) {
	listeners, err := c.store.ListListenersByNamespace(ctx, namespace)
	if err != nil {
		return nil, err
	}

	return store.FilterListenersByLabels(listeners.Items, labels), nil
}

func streamLogger(l logr.Logger, id int64) logr.Logger {
	return l.WithValues("stream", id)
}

func streamRequestLog(l logr.Logger, id int64, req *envoyapi.DiscoveryRequest) {
	requestLog(streamLogger(l, id), req)
}

func requestLogger(l logr.Logger, req *envoyapi.DiscoveryRequest) logr.Logger {
	return l.WithValues("current_version", req.VersionInfo, "node", req.GetNode().GetId())
}

func requestLog(l logr.Logger, req *envoyapi.DiscoveryRequest) {
	requestLogger(l, req).Info("request")
}
