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
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	apiv1 "github.com/110y/bootes/internal/k8s/api/v1"
	"github.com/110y/bootes/internal/k8s/internal/validator"
)

const (
	certDir         = "/tmp/k8s-webhook-server/serving-certs"
	healthzEndpoint = "/healthz"
	readyzEndpoint  = "/readyz"
	healthzName     = "healthz"
	readyzName      = "readyz"

	validatingWebhookPrefix    = "/validate-bootes-io-v1-"
	routeValidatingWebhookPath = validatingWebhookPrefix + "route"

	pprofEndpointPrefix  = "/debug/pprof"
	pprofIndexEndpoint   = pprofEndpointPrefix + "/"
	pprofCmdlineEndpoint = pprofEndpointPrefix + "/cmdlilne"
	pprofProfileEndpoint = pprofEndpointPrefix + "/profile"
	pprofSymbolEndpoint  = pprofEndpointPrefix + "/symbol"
	pprofTraceEndpoint   = pprofEndpointPrefix + "/trace"
)

var pprofHandlerMap = map[string]http.HandlerFunc{
	pprofIndexEndpoint:   pprof.Index,
	pprofCmdlineEndpoint: pprof.Cmdline,
	pprofProfileEndpoint: pprof.Profile,
	pprofSymbolEndpoint:  pprof.Symbol,
	pprofTraceEndpoint:   pprof.Trace,
}

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
		LeaderElection:         false,
		Port:                   c.WebhookServerPort,
		ReadinessEndpointName:  readyzEndpoint,
		LivenessEndpointName:   healthzEndpoint,
		CertDir:                certDir,
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

	webhookServer := manager.GetWebhookServer()
	webhookServer.Register(
		routeValidatingWebhookPath,
		&webhook.Admission{Handler: validator.NewRouteValidator(c.Logger.WithName("route_validator"))},
	)

	for endpoint, handler := range pprofHandlerMap {
		if err := manager.AddMetricsExtraHandler(endpoint, http.HandlerFunc(handler)); err != nil {
			return nil, fmt.Errorf("failed to register pprof handlers: %w", err)
		}
	}

	return manager, nil
}
