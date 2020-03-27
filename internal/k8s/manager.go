package k8s

import (
	"fmt"

	apiv1 "github.com/110y/bootes/internal/k8s/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
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
		return nil, err
	}

	manager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             s,
		MetricsBindAddress: fmt.Sprintf(":%d", c.MetricsServerPort),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %w", err)
	}

	return manager, nil
}
