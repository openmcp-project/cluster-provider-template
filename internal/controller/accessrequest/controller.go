package accessrequest

import (
	"context"

	"github.com/openmcp-project/controller-utils/pkg/clusters"
	clustersv1alpha1 "github.com/openmcp-project/openmcp-operator/api/clusters/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type AccessRequestReconciler struct {
	platformCluster *clusters.Cluster
	providerName    string
}

func NewAccessRequestReconciler(platformCluster *clusters.Cluster, providerName string) *AccessRequestReconciler {
	return &AccessRequestReconciler{
		platformCluster: platformCluster,
		providerName:    providerName,
	}
}

func (r *AccessRequestReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

func (r *AccessRequestReconciler) SetupWithManager(mgr manager.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clustersv1alpha1.AccessRequest{}).
		Complete(r)
}
