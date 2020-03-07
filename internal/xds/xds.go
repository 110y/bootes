package xds

import (
	"context"
	"fmt"
	"net"

	"github.com/110y/bootes/internal/k8s/store"
	"github.com/110y/bootes/internal/xds/cache"
	xdsgrpc "github.com/110y/bootes/internal/xds/internal/grpc"
	xdscache "github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/envoyproxy/go-control-plane/pkg/server"
	"google.golang.org/grpc"
)

type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
}

func NewServer(ctx context.Context, sc xdscache.SnapshotCache, c cache.Cache, s store.Store, config *Config) (*Server, error) {
	srv := server.NewServer(ctx, sc, &callbacks{
		xdsCache: c,
		k8sStore: s,
	})

	gc := &xdsgrpc.Config{
		EnableChannelz:   config.EnableGRPCChannelz,
		EnableReflection: config.EnableGRPCReflection,
	}
	gs := xdsgrpc.NewServer(ctx, srv, gc)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		// TODO: wrap
		return nil, err
	}

	return &Server{
		grpcServer: gs,
		listener:   lis,
	}, nil
}

func NewSnapshotCache() xdscache.SnapshotCache {
	// TODO: logger
	return xdscache.NewSnapshotCache(true, xdscache.IDHash{}, nil)
}

func (s *Server) Start() error {
	// TODO: wrap error
	return s.grpcServer.Serve(s.listener)
}
