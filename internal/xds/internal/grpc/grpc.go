package grpc

import (
	"context"

	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	xds "github.com/envoyproxy/go-control-plane/pkg/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func NewServer(ctx context.Context, snapshotCache cache.SnapshotCache) *grpc.Server {
	xs := xds.NewServer(ctx, snapshotCache, nil)
	gs := grpc.NewServer()

	discovery.RegisterAggregatedDiscoveryServiceServer(gs, xs)

	// TODO:
	reflection.Register(gs)

	return gs
}
