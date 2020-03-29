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
	ctx := context.Background()

	version := uuid.New().String()
	logger := r.logger.WithValues("version", version)

	logger.Info(fmt.Sprintf("Reconciling %s", req.NamespacedName))

	opts := []store.ListOption{}
	cluster, err := r.store.GetCluster(ctx, req.Name, req.Namespace)
	if err != nil {
		if !errors.Is(err, store.ErrNotFound) {
			logger.Error(err, "failed to get clusters")
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
			toNodeName(pod.Name, pod.Namespace),
			version,
			filterClusters(clusters.Items, pod.Labels),
		)
		if err != nil {
			logger.Error(err, "failed to update clusuters")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func filterClusters(clusters []*api.Cluster, podLabels map[string]string) []*api.Cluster {
	results := []*api.Cluster{}
	for _, c := range clusters {
		if matchSelector(c, podLabels) {
			results = append(results, c)
		}
	}

	return results
}
