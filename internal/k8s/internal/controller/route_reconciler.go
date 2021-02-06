package controller

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/110y/bootes/internal/k8s/store"
	"github.com/110y/bootes/internal/observer/trace"
	"github.com/110y/bootes/internal/xds/cache"
)

var _ reconcile.Reconciler = (*RouteReconciler)(nil)

func NewRouteReconciler(s store.Store, c cache.Cache, l logr.Logger) reconcile.Reconciler {
	return &RouteReconciler{
		store:  s,
		cache:  c,
		logger: l,
	}
}

type RouteReconciler struct {
	store  store.Store
	cache  cache.Cache
	logger logr.Logger
}

func (r *RouteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx, span := trace.NewSpan(ctx, "RouteReconciler.Reconcile")
	defer span.End()

	version := uuid.New().String()
	logger := r.logger.WithValues("version", version)

	logger.Info(fmt.Sprintf("Reconciling %s", req.NamespacedName))

	opts := []store.ListOption{}
	route, err := r.store.GetRoute(ctx, req.Name, req.Namespace)
	if err != nil {
		if !errors.Is(err, store.ErrNotFound) {
			logger.Error(err, "failed to get route")
			return ctrl.Result{}, err
		}
	} else {
		if route.Spec.WorkloadSelector != nil {
			opts = append(opts, store.WithLabelFilter(route.Spec.WorkloadSelector.Labels))
		}
	}

	pods, err := r.store.ListPodsByNamespace(ctx, req.Namespace, opts...)
	if err != nil {
		logger.Error(err, "failed to list pods")
		return ctrl.Result{}, err
	}

	routes, err := r.store.ListRoutesByNamespace(ctx, req.Namespace)
	if err != nil {
		logger.Error(err, "failed to list routes")
		return ctrl.Result{}, err
	}

	for _, pod := range pods.Items {
		err := r.cache.UpdateRoutes(
			ctx,
			store.ToNodeName(pod.Name, pod.Namespace),
			version,
			store.FilterRoutesByLabels(routes.Items, pod.Labels),
		)
		if err != nil {
			logger.Error(err, "failed to update clusuters")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}
