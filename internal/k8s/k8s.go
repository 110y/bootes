package k8s

import (
	"fmt"

	apiv1 "github.com/110y/bootes/internal/k8s/api/v1"
	"github.com/110y/bootes/internal/k8s/internal/controller"
	"github.com/110y/bootes/internal/k8s/store"
	"github.com/110y/bootes/internal/xds/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type Controller struct {
	manager manager.Manager
}

func NewController(mgr manager.Manager, s store.Store, c cache.Cache) (*Controller, error) {
	if err := setupClusterReconciler(mgr, s, c); err != nil {
		// TODO:
		return nil, err
	}

	return &Controller{manager: mgr}, nil
}

func setupClusterReconciler(mgr manager.Manager, s store.Store, c cache.Cache) error {
	cr := controller.NewClusterReconciler(s, c)

	if err := ctrl.NewControllerManagedBy(mgr).For(&apiv1.Cluster{}).Complete(cr); err != nil {
		return fmt.Errorf("failed to setup cluster reconciler: %s\n", err)
	}

	return nil
}

func (c *Controller) Start() error {
	// TODO: do not use signal handler directly
	return c.manager.Start(ctrl.SetupSignalHandler())
}
