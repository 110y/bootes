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
	os.Exit(run(context.Background()))
}

func run(ctx context.Context) int {
	l, err := newLogger()
	if err != nil {
		_, ferr := fmt.Fprintf(os.Stderr, "failed to create logger: %s", err)
		if ferr != nil {
			// Unhandleable, something went wrong...
			panic(fmt.Sprintf("failed to write log:`%s` original error is:`%s`", ferr, err))
		}
	}

	sl := l.WithName("server")

	xl := l.WithName("xds")
	sc := xds.NewSnapshotCache(xl.WithName("snapshot_cache"))
	c := cache.New(sc)

	mgr, err := k8s.NewManager()
	if err != nil {
		sl.Error(err, "failed to create k8s manager")
		return 1
	}

	s := store.New(mgr.GetClient(), mgr.GetAPIReader())

	env, err := getEnvironments()
	if err != nil {
		sl.Error(err, "failed to load environment variables")
		return 1
	}

	xs, err := xds.NewServer(ctx, sc, c, s, xl, &xds.Config{
		Port:                 env.XDSGRPCPort,
		EnableGRPCChannelz:   env.XDSGRPCEnableChannelz,
		EnableGRPCReflection: env.XDSGRPCEnableReflection,
	})
	if err != nil {
		sl.Error(err, "failed to create xds server")
		return 1
	}

	ctrl, err := k8s.NewController(mgr, s, c, l.WithName("k8s"))
	if err != nil {
		sl.Error(err, "failed to create k8s controller")
		return 1
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
		return 0
	case <-errChan:
		// TODO:
		sl.Error(err, "failed to run server")
		return 1
	}
}
