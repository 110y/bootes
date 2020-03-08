package xds

import (
	"context"
	"fmt"

	"github.com/110y/bootes/internal/k8s/store"
	"github.com/110y/bootes/internal/xds/cache"
	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/pkg/server"
	"github.com/go-logr/logr"
)

var _ server.Callbacks = &callbacks{}

type callbacks struct {
	cache  cache.Cache
	store  store.Store
	logger logr.Logger
}

func (c *callbacks) OnStreamOpen(context.Context, int64, string) error {
	fmt.Println("OnStreamOpen")

	return nil
}

func (c *callbacks) OnStreamClosed(int64) {
	fmt.Println("OnStreamClosed")
}

func (c *callbacks) OnStreamRequest(streamID int64, req *api.DiscoveryRequest) error {
	fmt.Println(fmt.Sprintf("OnStreamRequest. version:`%s`", req.VersionInfo))

	// TODO: get clusters by namespace

	// TODO: filter clusters by label selector

	// TODO: save clusters to cache

	// TODO: set cache to node

	return nil
}

func (c *callbacks) OnStreamResponse(streamID int64, req *api.DiscoveryRequest, res *api.DiscoveryResponse) {
	fmt.Println(fmt.Sprintf("OnStreamResponse. version:`%s`", res.VersionInfo))
}

func (c callbacks) OnFetchRequest(context.Context, *api.DiscoveryRequest) error {
	fmt.Println("OnFetchRequest")
	return nil
}

func (c *callbacks) OnFetchResponse(*api.DiscoveryRequest, *api.DiscoveryResponse) {
	fmt.Println("OnFetchResponse")
}
