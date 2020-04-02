package xds

import (
	"context"

	"github.com/110y/bootes/internal/k8s/store"
	"github.com/110y/bootes/internal/xds/cache"
	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/pkg/server"
	"github.com/go-logr/logr"
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
	streamLogger(c.loggerOnStreamOpen, streamID).Info("closed")
}

func (c *callbacks) OnStreamRequest(streamID int64, req *api.DiscoveryRequest) error {
	streamRequestLog(c.loggerOnStreamRequest, streamID, req)

	// TODO: get cache by node

	// TODO: get clusters by namespace

	// TODO: filter clusters by label selector

	// TODO: save clusters to cache

	// TODO: set cache to node

	return nil
}

func (c *callbacks) OnStreamResponse(streamID int64, req *api.DiscoveryRequest, _ *api.DiscoveryResponse) {
	streamRequestLog(c.loggerOnStreamResponse, streamID, req)
}

func (c callbacks) OnFetchRequest(_ context.Context, req *api.DiscoveryRequest) error {
	requestLog(c.loggerOnFetchRequest, req)
	return nil
}

func (c *callbacks) OnFetchResponse(req *api.DiscoveryRequest, _ *api.DiscoveryResponse) {
	requestLog(c.loggerOnFetchResponse, req)
}

func streamLogger(l logr.Logger, id int64) logr.Logger {
	return l.WithValues("stream", id)
}

func streamRequestLog(l logr.Logger, id int64, req *api.DiscoveryRequest) {
	requestLog(streamLogger(l, id), req)
}

func requestLog(l logr.Logger, req *api.DiscoveryRequest) {
	l.Info("request", "version", req.VersionInfo, "node", req.GetNode().GetId())
}
