package controller

import (
	"context"
	"fmt"

	api "github.com/110y/bootes/internal/k8s/api/v1"
	"github.com/110y/bootes/internal/k8s/store"
	"github.com/110y/bootes/internal/xds/cache"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewClusterReconciler(s store.Store, c cache.Cache, l logr.Logger) reconcile.Reconciler {
	return &ClusterReconciler{
		store:  s,
		cache:  c,
		logger: l,
		// Scheme: manager.GetScheme(),
	}
}

type ClusterReconciler struct {
	store  store.Store
	cache  cache.Cache
	logger logr.Logger
	// Scheme *runtime.Scheme
}

func (r *ClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	cluster, err := r.store.GetCluster(ctx, req.Name, req.Namespace)
	if err != nil {
		// TODO: treat not found
		fmt.Println(fmt.Sprintf("RECONCILE: ERROR: %s", err))
		return ctrl.Result{}, err
	}

	// TODO: list pods by namespace
	// podList, err := r.store.ListPodsByNamespace(ctx, req.Namespace)
	// if err != nil {
	//     // TODO: handle err
	//     return ctrl.Result{}, err
	// }

	// for _, pod := range podList.Items {
	// }

	// TODO: add cluster to each pod

	r.logger.Info(fmt.Sprintf("RECONCILE: %s", cluster.Spec.Name))

	if err := r.cache.AddCluster("id", cluster); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).For(&api.Cluster{}).Complete(r)
}
