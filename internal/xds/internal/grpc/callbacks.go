package grpc

import (
	"context"
	"fmt"

	api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/pkg/server"
)

type callbacks struct{}

func (c *callbacks) OnStreamOpen(context.Context, int64, string) error {
	fmt.Println("OnStreamOpen")
	return nil
}

func (c *callbacks) OnStreamClosed(int64) {
	fmt.Println("OnStreamClosed")
}

func (c *callbacks) OnStreamRequest(int64, *api.DiscoveryRequest) error {
	fmt.Println("OnStreamRequest")
	return nil
}

func (c *callbacks) OnStreamResponse(int64, *api.DiscoveryRequest, *api.DiscoveryResponse) {
	fmt.Println("OnStreamResponse")
}

func (c callbacks) OnFetchRequest(context.Context, *api.DiscoveryRequest) error {
	fmt.Println("OnFetchRequest")
	return nil
}

func (c *callbacks) OnFetchResponse(*api.DiscoveryRequest, *api.DiscoveryResponse) {
	fmt.Println("OnFetchResponse")
}

var _ server.Callbacks = &callbacks{}
