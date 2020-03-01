package grpc

import (
	"context"

	"github.com/110y/bootes/internal/k8s/store"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	xds "github.com/envoyproxy/go-control-plane/pkg/server"
	"google.golang.org/grpc"
	channelz "google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/reflection"
)

func NewServer(ctx context.Context, snapshotCache cache.SnapshotCache, k8sStore store.Store, config *Config) *grpc.Server {
	xs := xds.NewServer(ctx, snapshotCache, &callbacks{k8sStore: k8sStore})
	gs := grpc.NewServer()

	discovery.RegisterAggregatedDiscoveryServiceServer(gs, xs)

	if config.EnableChannelz {
		channelz.RegisterChannelzServiceToServer(gs)
	}

	if config.EnableReflection {
		reflection.Register(gs)
	}

	return gs
}
