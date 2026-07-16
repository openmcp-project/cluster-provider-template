package cluster

import (
	"context"

	"github.com/openmcp-project/controller-utils/pkg/clusters"
	clustersv1alpha1 "github.com/openmcp-project/openmcp-operator/api/clusters/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ClusterReconciler struct {
	platformCluster *clusters.Cluster
	providerName    string
}

func NewClusterReconciler(platformCluster *clusters.Cluster, providerName string) *ClusterReconciler {
	return &ClusterReconciler{
		platformCluster: platformCluster,
		providerName:    providerName,
	}
}

func (r *ClusterReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

func (r *ClusterReconciler) SetupWithManager(mgr manager.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clustersv1alpha1.Cluster{}).
		Complete(r)
}
