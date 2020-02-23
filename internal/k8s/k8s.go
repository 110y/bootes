package k8s

import (
	"github.com/110y/bootes/internal/cache"
	"github.com/110y/bootes/internal/k8s/internal/controller"
)

type Controller struct {
	controller *controller.Controller
	cache      *cache.Cache
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
		cache:      sc,
	}, nil
}

func (c *Controller) Start() error {
	return c.controller.Start()
}
