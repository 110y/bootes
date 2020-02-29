package controller

import (
	"fmt"

	"github.com/110y/bootes/internal/cache"
	apiv1 "github.com/110y/bootes/internal/k8s/api/v1"
	"github.com/110y/bootes/internal/k8s/store"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Controller struct {
	manager ctrl.Manager
	store   store.Store
	cache   *cache.Cache
}

func NewController(cache *cache.Cache) (*Controller, error) {
	s := runtime.NewScheme()
	if err := scheme.AddToScheme(s); err != nil {
		return nil, fmt.Errorf("failed to create new scheme: %s\n", err)
	}

	if err := apiv1.AddToScheme(s); err != nil {
		return nil, fmt.Errorf("failed to add scheme to apiv1: %s\n", err)
	}

	cfg, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}

	manager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: s,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %s\n", err)
	}

	st := store.NewStore(manager.GetClient())

	cr := &ClusterReconciler{
		store: st,
		cache: cache,
		// Scheme: manager.GetScheme(),
	}

	if err := cr.SetupWithManager(manager); err != nil {
		return nil, fmt.Errorf("failed to setup manager: %s\n", err)
	}

	return &Controller{
		manager: manager,
		store:   st,
		cache:   cache,
	}, nil
}

func (c *Controller) Start() error {
	// TODO: do not use signal handler directly
	return c.manager.Start(ctrl.SetupSignalHandler())
}

func (c *Controller) GetStore() store.Store {
	return c.store
}
