package controller

import (
	"context"
	"errors"
	"fmt"

	"github.com/110y/bootes/internal/k8s/store"
	"github.com/110y/bootes/internal/observer/trace"
	"github.com/110y/bootes/internal/xds/cache"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ reconcile.Reconciler = (*ListenerReconciler)(nil)

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
	ctx, span := trace.NewSpan(context.Background(), "ListenerReconciler.Reconcile")
	defer span.End()

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
		logger.Error(err, "failed to list listeners")
		return ctrl.Result{}, err
	}

	for _, pod := range pods.Items {
		err := r.cache.UpdateListeners(
			ctx,
			store.ToNodeName(pod.Name, pod.Namespace),
			version,
			store.FilterListenersByLabels(listeners.Items, pod.Labels),
		)
		if err != nil {
			logger.Error(err, "failed to update clusuters")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}
