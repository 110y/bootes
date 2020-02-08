package xds

import (
	"context"
	"fmt"
	"net"

	"github.com/110y/bootes/internal/cache"
	xdsgrpc "github.com/110y/bootes/internal/xds/internal/grpc"
	"google.golang.org/grpc"
)

type Server struct {
	grpcServer    *grpc.Server
	listener      net.Listener
	snapshotCache *cache.Cache
}

func NewServer(ctx context.Context, snapshotCache *cache.Cache, config *Config) (*Server, error) {
	gc := &xdsgrpc.Config{
		EnableChannelz:   config.EnableGRPCChannelz,
		EnableReflection: config.EnableGRPCReflection,
	}
	gs := xdsgrpc.NewServer(ctx, snapshotCache, gc)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		// TODO: wrap
		return nil, err
	}

	return &Server{
		grpcServer:    gs,
		listener:      lis,
		snapshotCache: snapshotCache,
	}, nil
}

func (s *Server) Start() error {
	return s.grpcServer.Serve(s.listener)
}
