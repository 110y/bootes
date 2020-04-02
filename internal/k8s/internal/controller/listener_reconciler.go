package controller

import (
	"context"
	"errors"
	"fmt"

	api "github.com/110y/bootes/internal/k8s/api/v1"
	"github.com/110y/bootes/internal/k8s/store"
	"github.com/110y/bootes/internal/xds/cache"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewListenerReconciler(s store.Store, c cache.Cache, l logr.Logger) reconcile.Reconciler {
	return &ListenerReconciler{
		store:  s,
		cache:  c,
		logger: l,
	}
}

type ListenerReconciler struct {
	store  store.Store
	cache  cache.Cache
	logger logr.Logger
}

func (r *ListenerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	version := uuid.New().String()
	logger := r.logger.WithValues("version", version)

	logger.Info(fmt.Sprintf("Reconciling %s", req.NamespacedName))

	opts := []store.ListOption{}
	listener, err := r.store.GetListener(ctx, req.Name, req.Namespace)
	if err != nil {
		if !errors.Is(err, store.ErrNotFound) {
			logger.Error(err, "failed to get listener")
			return ctrl.Result{}, err
		}
	} else {
		if listener.Spec.WorkloadSelector != nil {
			opts = append(opts, store.WithLabelFilter(listener.Spec.WorkloadSelector.Labels))
		}
	}

	pods, err := r.store.ListPodsByNamespace(ctx, req.Namespace, opts...)
	if err != nil {
		logger.Error(err, "failed to list pods")
		return ctrl.Result{}, err
	}

	listeners, err := r.store.ListListenersByNamespace(ctx, req.Namespace)
	if err != nil {
		logger.Error(err, "failed to list clusters")
		return ctrl.Result{}, err
	}

	for _, pod := range pods.Items {
		err := r.cache.UpdateListeners(
			toNodeName(pod.Name, pod.Namespace),
			version,
			filterListeners(listeners.Items, pod.Labels),
		)
		if err != nil {
			logger.Error(err, "failed to update clusuters")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func filterListeners(listeners []*api.Listener, podLabels map[string]string) []*api.Listener {
	results := []*api.Listener{}
	for _, l := range listeners {
		if matchSelector(l, podLabels) {
			results = append(results, l)
		}
	}

	return results
}