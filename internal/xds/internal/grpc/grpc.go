package grpc

import (
	"context"

	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	xds "github.com/envoyproxy/go-control-plane/pkg/server"
	"google.golang.org/grpc"
	channelz "google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/reflection"
)

func NewServer(ctx context.Context, snapshotCache cache.SnapshotCache, config *Config) *grpc.Server {
	xs := xds.NewServer(ctx, snapshotCache, &callbacks{})
	gs := grpc.NewServer()

	discovery.RegisterAggregatedDiscoveryServiceServer(gs, xs)
	// api.RegisterClusterDiscoveryServiceServer(gs, xs)

	if config.EnableChannelz {
		channelz.RegisterChannelzServiceToServer(gs)
	}

	if config.EnableReflection {
		reflection.Register(gs)
	}

	return gs
}
