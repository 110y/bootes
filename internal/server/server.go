package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/110y/servergroup"
	"golang.org/x/sys/unix"

	"github.com/110y/bootes/internal/k8s"
	"github.com/110y/bootes/internal/k8s/store"
	"github.com/110y/bootes/internal/observer/trace"
	"github.com/110y/bootes/internal/xds"
	"github.com/110y/bootes/internal/xds/cache"
)

func Run(ctx context.Context) int {
	ctx, stop := signal.NotifyContext(ctx, unix.SIGTERM, unix.SIGINT)
	defer stop()

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

	var sg servergroup.Group
	sg.Add(ctrl)
	sg.Add(xs)

	if err := sg.Start(ctx); err != nil {
		sl.Error(err, "failed to start or stop servers")
		return 1
	}

	sl.Info("succeeded to terminate servers")
	return 0
}
