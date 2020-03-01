package k8s

import (
	"github.com/110y/bootes/internal/cache"
	"github.com/110y/bootes/internal/k8s/internal/controller"
	"github.com/110y/bootes/internal/k8s/store"
)

type Controller struct {
	controller *controller.Controller
}

func NewController(sc *cache.Cache) (*Controller, error) {
	// TODO:
	ctrl, err := controller.NewController(sc)
	if err != nil {
		// TODO:
		return nil, err
	}

	return &Controller{
		controller: ctrl,
	}, nil
}

func (c *Controller) Start() error {
	return c.controller.Start()
}

func (c *Controller) GetStore() store.Store {
	return c.controller.GetStore()
}
