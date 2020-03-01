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

	controller, err := k8s.NewController(c)
	if err != nil {
		// TODO:
		return err
	}

	xs, err := xds.NewServer(ctx, c, controller.GetStore(), &xds.Config{
		Port:                 8081, // TODO:
		EnableGRPCChannelz:   true, // TODO:
		EnableGRPCReflection: true, // TODO:
	})
	if err != nil {
		// TODO: wrap
		return err
	}

	errChan := make(chan error, 1)

	go func() {
		if err := xs.Start(); err != nil {
			errChan <- err
		}
	}()

	go func() {
		if err := controller.Start(); err != nil {
			errChan <- err
		}
	}()

	terminationChan := make(chan os.Signal, 1)
	signal.Notify(terminationChan, syscall.SIGTERM, syscall.SIGINT)

	// TODO:
	select {
	case <-terminationChan:
		// TODO: stop servers
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

	os.Exit(0)
}
