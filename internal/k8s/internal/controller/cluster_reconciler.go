package controller

import (
	"context"

	api "github.com/110y/bootes/internal/k8s/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterReconciler struct {
	client.Client
	// Scheme *runtime.Scheme
}

func (r *ClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	var cluster *api.Cluster
	if err := r.Get(ctx, req.NamespacedName, cluster); err != nil {
		// TODO:
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).For(&api.Cluster{}).Complete(r)
}
