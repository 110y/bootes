package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/110y/bootes/internal/k8s"
	"github.com/110y/bootes/internal/k8s/store"
	"github.com/110y/bootes/internal/xds"
	"github.com/110y/bootes/internal/xds/cache"
)

func Run() {
	exit(run(context.Background()))
}

func exit(err error) {
	if err != nil {
		// TODO: implement
		fmt.Fprintf(os.Stderr, err.Error())

		os.Exit(1)
	}

	os.Exit(0)
}

func run(ctx context.Context) error {
	sc := xds.NewSnapshotCache()
	c := cache.New(sc)

	mgr, err := k8s.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create k8s manager: %w", err)
	}

	s := store.New(mgr.GetClient(), mgr.GetAPIReader())

	xs, err := xds.NewServer(ctx, sc, c, s, &xds.Config{
		Port:                 8081, // TODO:
		EnableGRPCChannelz:   true, // TODO:
		EnableGRPCReflection: true, // TODO:
	})
	if err != nil {
		return fmt.Errorf("failed to create xds server: %w", err)
	}

	ctrl, err := k8s.NewController(mgr, s, c)
	if err != nil {
		// TODO:
		return err
	}

	errChan := make(chan error, 1)

	go func() {
		if err := xs.Start(); err != nil {
			errChan <- err
		}
	}()

	go func() {
		if err := ctrl.Start(); err != nil {
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
