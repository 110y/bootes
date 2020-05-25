package k8s

import (
	"fmt"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	apiv1 "github.com/110y/bootes/internal/k8s/api/v1"
	"github.com/110y/bootes/internal/k8s/internal/controller"
	"github.com/110y/bootes/internal/k8s/store"
	"github.com/110y/bootes/internal/xds/cache"
)

type Controller struct {
	manager manager.Manager
	logger  logr.Logger
}

func NewController(mgr manager.Manager, s store.Store, c cache.Cache, l logr.Logger) (*Controller, error) {
	ctrl.SetLogger(l)

	if err := setupClusterReconciler(mgr, s, c, l.WithName("cluster_reconciler")); err != nil {
		return nil, err
	}

	if err := setupListenerReconciler(mgr, s, c, l.WithName("listener_reconciler")); err != nil {
		return nil, err
	}

	if err := setupRouteReconciler(mgr, s, c, l.WithName("route_reconciler")); err != nil {
		return nil, err
	}

	if err := setupEndpointReconciler(mgr, s, c, l.WithName("endpoint_reconciler")); err != nil {
		return nil, err
	}

	return &Controller{
		manager: mgr,
		logger:  l,
	}, nil
}

func setupClusterReconciler(mgr manager.Manager, s store.Store, c cache.Cache, l logr.Logger) error {
	cr := controller.NewClusterReconciler(s, c, l)

	if err := ctrl.NewControllerManagedBy(mgr).For(&apiv1.Cluster{}).Complete(cr); err != nil {
		return fmt.Errorf("failed to setup cluster reconciler: %s", err)
	}

	return nil
}

func setupListenerReconciler(mgr manager.Manager, s store.Store, c cache.Cache, l logr.Logger) error {
	lr := controller.NewListenerReconciler(s, c, l)

	if err := ctrl.NewControllerManagedBy(mgr).For(&apiv1.Listener{}).Complete(lr); err != nil {
		return fmt.Errorf("failed to setup listener reconciler: %s", err)
	}

	return nil
}

func setupRouteReconciler(mgr manager.Manager, s store.Store, c cache.Cache, l logr.Logger) error {
	rr := controller.NewRouteReconciler(s, c, l)

	if err := ctrl.NewControllerManagedBy(mgr).For(&apiv1.Route{}).Complete(rr); err != nil {
		return fmt.Errorf("failed to setup route reconciler: %s", err)
	}

	return nil
}

func setupEndpointReconciler(mgr manager.Manager, s store.Store, c cache.Cache, l logr.Logger) error {
	rr := controller.NewEndpointReconciler(s, c, l)

	if err := ctrl.NewControllerManagedBy(mgr).For(&apiv1.Endpoint{}).Complete(rr); err != nil {
		return fmt.Errorf("failed to setup endpoint reconciler: %s", err)
	}

	return nil
}

func (c *Controller) Start(stopCh chan struct{}) error {
	c.logger.Info("starting k8s controller")
	return c.manager.Start(stopCh)
}
