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

var _ reconcile.Reconciler = (*ClusterReconciler)(nil)

func NewClusterReconciler(s store.Store, c cache.Cache, l logr.Logger) reconcile.Reconciler {
	return &ClusterReconciler{
		store:  s,
		cache:  c,
		logger: l,
	}
}

type ClusterReconciler struct {
	store  store.Store
	cache  cache.Cache
	logger logr.Logger
}

func (r *ClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx, span := trace.NewSpan(context.Background(), "ClusterReconciler.Reconcile")
	defer span.End()

	version := uuid.New().String()
	logger := r.logger.WithValues("version", version)

	logger.Info(fmt.Sprintf("Reconciling %s", req.NamespacedName))

	opts := []store.ListOption{}
	cluster, err := r.store.GetCluster(ctx, req.Name, req.Namespace)
	if err != nil {
		if !errors.Is(err, store.ErrNotFound) {
			logger.Error(err, "failed to get cluster")
			return ctrl.Result{}, err
		}
	} else {
		if cluster.Spec.WorkloadSelector != nil {
			opts = append(opts, store.WithLabelFilter(cluster.Spec.WorkloadSelector.Labels))
		}
	}

	pods, err := r.store.ListPodsByNamespace(ctx, req.Namespace, opts...)
	if err != nil {
		logger.Error(err, "failed to list pods")
		return ctrl.Result{}, err
	}

	clusters, err := r.store.ListClustersByNamespace(ctx, req.Namespace)
	if err != nil {
		logger.Error(err, "failed to list clusters")
		return ctrl.Result{}, err
	}

	for _, pod := range pods.Items {
		err := r.cache.UpdateClusters(
			ctx,
			store.ToNodeName(pod.Name, pod.Namespace),
			version,
			store.FilterClustersByLabels(clusters.Items, pod.Labels),
		)
		if err != nil {
			logger.Error(err, "failed to update clusuters")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}
