package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/110y/bootes/internal/cache"
	"github.com/110y/bootes/internal/k8s"
	"github.com/110y/bootes/internal/xds"
)

func Run() {
	exit(run(context.Background()))
}

func run(ctx context.Context) error {
	c := cache.NewCache()

	xs, err := xds.NewServer(ctx, c, &xds.Config{
		Port:                 8080, // TODO:
		EnableGRPCChannelz:   true, // TODO:
		EnableGRPCReflection: true, // TODO:
	})
	if err != nil {
		// TODO: wrap
		return err
	}

	k8s.NewController(c)

	errChan := make(chan error, 1)

	// TODO:
	go func() {
		if err := xs.Start(); err != nil {
			errChan <- err
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-sigChan:
		// TODO: stop
		return nil
	case <-errChan:
		return err
	}
}

func exit(err error) {
	if err != nil {
		// TODO: implement
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}
