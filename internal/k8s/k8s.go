package k8s

import (
	"fmt"

	apiv1 "github.com/110y/bootes/internal/k8s/api/v1"
	"github.com/110y/bootes/internal/k8s/internal/controller"
	"github.com/110y/bootes/internal/k8s/store"
	"github.com/110y/bootes/internal/xds/cache"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type Controller struct {
	manager manager.Manager
}

func NewController(mgr manager.Manager, s store.Store, c cache.Cache, l logr.Logger) (*Controller, error) {
	ctrl.SetLogger(l)

	if err := setupClusterReconciler(mgr, s, c, l.WithName("cluster_reconciler")); err != nil {
		return nil, err
	}

	return &Controller{manager: mgr}, nil
}

func setupClusterReconciler(mgr manager.Manager, s store.Store, c cache.Cache, l logr.Logger) error {
	cr := controller.NewClusterReconciler(s, c, l)

	if err := ctrl.NewControllerManagedBy(mgr).For(&apiv1.Cluster{}).Complete(cr); err != nil {
		return fmt.Errorf("failed to setup cluster reconciler: %s", err)
	}

	return nil
}

func (c *Controller) Start() error {
	// TODO: do not use signal handler directly
	return c.manager.Start(ctrl.SetupSignalHandler())
}
