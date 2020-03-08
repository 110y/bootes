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
	"github.com/go-logr/logr"
	"google.golang.org/grpc"
)

type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
}

func NewServer(ctx context.Context, sc xdscache.SnapshotCache, c cache.Cache, s store.Store, l logr.Logger, config *Config) (*Server, error) {
	srv := server.NewServer(ctx, sc, &callbacks{
		cache:  c,
		store:  s,
		logger: l.WithName("callbacks"),
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

func NewSnapshotCache(l logr.Logger) xdscache.SnapshotCache {
	return xdscache.NewSnapshotCache(true, xdscache.IDHash{}, newSnapshotCacheLogger(l))
}

func (s *Server) Start() error {
	// TODO: wrap error
	return s.grpcServer.Serve(s.listener)
}
