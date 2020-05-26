package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/110y/bootes/internal/k8s"
	"github.com/110y/bootes/internal/k8s/store"
	"github.com/110y/bootes/internal/observer/trace"
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
		return 1
	}

	sl := l.WithName("server")

	env, err := getEnvironments()
	if err != nil {
		sl.Error(err, "failed to load environment variables")
		return 1
	}

	flush, err := trace.Initialize(&trace.Config{
		UseStdout:              env.TraceUseStdout,
		UseJaeger:              env.TraceUseJaeger,
		JaegerEndpoint:         env.TraceJaegerEndpoint,
		UseGCPCloudTrace:       env.TraceUseGCPCloudTrace,
		GCPCloudTraceProjectID: env.TraceGCPCloudTraceProjectID,
		Logger:                 l.WithName("trace"),
	})
	if err != nil {
		sl.Error(err, "failed to initialize tracer")
		return 1
	}
	defer flush()

	xl := l.WithName("xds")
	sc := xds.NewSnapshotCache(xl.WithName("snapshot_cache"))
	c := cache.New(sc)

	mgr, err := k8s.NewManager(&k8s.ManagerConfig{
		HealthzServerPort: env.HealthProbeServerPort,
		MetricsServerPort: env.K8SMetricsServerPort,
	})
	if err != nil {
		sl.Error(err, "failed to create k8s manager")
		return 1
	}

	s := store.New(mgr.GetClient(), mgr.GetAPIReader())

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

	xdsErrChan := make(chan error, 1)
	xdsStopChan := make(chan struct{}, 1)
	go func() {
		xdsErrChan <- xs.Start(xdsStopChan)
	}()

	k8sErrChan := make(chan error, 1)
	k8sStopChan := make(chan struct{}, 1)
	go func() {
		k8sErrChan <- ctrl.Start(k8sStopChan)
	}()

	terminationChan := make(chan os.Signal, 1)
	signal.Notify(terminationChan, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-terminationChan:
		sl.Info("stopping servers")

		xdsStopChan <- struct{}{}
		k8sStopChan <- struct{}{}

		wg := sync.WaitGroup{}

		wg.Add(1)
		isFailedStopXDSServer := false
		go func() {
			defer wg.Done()
			err := <-xdsErrChan
			if err != nil {
				sl.Error(err, "xds grpc server did not stop correctly after termination signal received")
				isFailedStopXDSServer = true
			}
			sl.Info("xds server has stopped")
		}()

		wg.Add(1)
		isFailedStopK8SController := false
		go func() {
			defer wg.Done()
			err := <-k8sErrChan
			if err != nil {
				sl.Error(err, "k8s controller did not stop correctly after termination signal received")
				isFailedStopK8SController = true
			}
			sl.Info("k8s controller has stopped")
		}()

		wg.Wait()

		if isFailedStopXDSServer || isFailedStopK8SController {
			return 1
		}

		sl.Info("succeeded to shut down server")
		return 0

	case err := <-xdsErrChan:
		sl.Error(err, "failed to run xds grpc server")
		k8sStopChan <- struct{}{}

		nerr := <-k8sErrChan
		if nerr != nil {
			sl.Error(nerr, "failed to stop k8s controller after xds grpc server returned error")
		}

		return 1
	case err := <-k8sErrChan:
		sl.Error(err, "failed to run k8s controller")
		xdsStopChan <- struct{}{}

		nerr := <-xdsErrChan
		if nerr != nil {
			sl.Error(nerr, "failed to stop xds grpc server after k8s controller returned error")
		}

		return 1
	}
}
