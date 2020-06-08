package grpc

import (
	"context"

	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"google.golang.org/grpc"
	channelz "google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/reflection"
)

func NewServer(ctx context.Context, xs discovery.AggregatedDiscoveryServiceServer, config *Config) *grpc.Server {
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
