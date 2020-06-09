package k8s

import (
	"fmt"
	"net/http"
	"net/http/pprof"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	apiv1 "github.com/110y/bootes/internal/k8s/api/v1"
)

const (
	healthzEndpoint = "/healthz"
	readyzEndpoint  = "/readyz"
	healthzName     = "healthz"
	readyzName      = "readyz"

	pprofEndpointPrefix  = "/debug/pprof"
	pprofIndexEndpoint   = pprofEndpointPrefix + "/"
	pprofCmdlineEndpoint = pprofEndpointPrefix + "/cmdlilne"
	pprofProfileEndpoint = pprofEndpointPrefix + "/profile"
	pprofSymbolEndpoint  = pprofEndpointPrefix + "/symbol"
	pprofTraceEndpoint   = pprofEndpointPrefix + "/trace"
)

func NewManager(c *ManagerConfig) (manager.Manager, error) {
	s := runtime.NewScheme()
	if err := scheme.AddToScheme(s); err != nil {
		return nil, fmt.Errorf("failed to create new scheme: %w", err)
	}
	if err := apiv1.AddToScheme(s); err != nil {
		return nil, fmt.Errorf("failed to add scheme to apiv1: %w", err)
	}

	cfg, err := ctrl.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubernetes configuration: %w", err)
	}

	manager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:                 s,
		ReadinessEndpointName:  readyzEndpoint,
		LivenessEndpointName:   healthzEndpoint,
		HealthProbeBindAddress: fmt.Sprintf(":%d", c.HealthzServerPort),
		MetricsBindAddress:     fmt.Sprintf(":%d", c.MetricsServerPort),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %w", err)
	}

	if err := manager.AddHealthzCheck(healthzName, healthz.Ping); err != nil {
		return nil, fmt.Errorf("failed to register healthz checker: %w", err)
	}

	if err := manager.AddReadyzCheck(readyzName, healthz.Ping); err != nil {
		return nil, fmt.Errorf("failed to register readyz checker: %w", err)
	}

	if err := setPprofHandlelrs(manager); err != nil {
		return nil, fmt.Errorf("failed to register pprof handlers: %w", err)
	}

	return manager, nil
}

func setPprofHandlelrs(mgr manager.Manager) error {
	if err := mgr.AddMetricsExtraHandler(pprofIndexEndpoint, http.HandlerFunc(pprof.Index)); err != nil {
		return err
	}

	if err := mgr.AddMetricsExtraHandler(pprofCmdlineEndpoint, http.HandlerFunc(pprof.Cmdline)); err != nil {
		return err
	}

	if err := mgr.AddMetricsExtraHandler(pprofProfileEndpoint, http.HandlerFunc(pprof.Profile)); err != nil {
		return err
	}

	if err := mgr.AddMetricsExtraHandler(pprofSymbolEndpoint, http.HandlerFunc(pprof.Symbol)); err != nil {
		return err
	}

	if err := mgr.AddMetricsExtraHandler(pprofTraceEndpoint, http.HandlerFunc(pprof.Trace)); err != nil {
		return err
	}

	return nil
}
