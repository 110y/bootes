package k8s

import (
	"fmt"

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

	return manager, nil
}
