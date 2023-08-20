package controller

import (
	"context"
	"time"

	v1 "jiayi.com/auth-center/api/v1"
	"jiayi.com/auth-center/pkg/config"
	"jiayi.com/auth-center/pkg/service"
	"jiayi.com/auth-center/pkg/util"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type RoleBindingReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *RoleBindingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("traceID", time.Now().UnixNano())
	logger.Info("RoleBindingReconciler.Reconcile start")
	rb := new(rbacv1.RoleBinding)
	err := r.Get(ctx, req.NamespacedName, rb)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			logger.Error(err, "RoleBindingReconciler.Reconcile get rolebinding error")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		return ctrl.Result{}, nil
	}
	if !rb.DeletionTimestamp.IsZero() {
		err := service.NewK8sService(ctx, logger, r.Client).RemoveRoleBindingOfAuthCenterByRoleBinding(rb)
		if err != nil {
			logger.Error(err, "RoleBindingReconciler.Reconcile remove rolebinding error")
			return ctrl.Result{}, err
		}
		rb.ObjectMeta.Finalizers = util.RemoveString(rb.ObjectMeta.Finalizers, config.FinalizerNsAuthCenter)
		err = r.Update(ctx, rb)
		return ctrl.Result{}, err
	}
	logger.Info("RoleBindingReconciler.Reconcile end")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RoleBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rbacv1.RoleBinding{}, builder.WithPredicates(predicate.NewPredicateFuncs(func(o client.Object) bool {
			log.Log.Info(o.GetName())
			for _, ownerReference := range o.GetOwnerReferences() {
				if ownerReference.Kind == v1.Kind {
					return true
				}
			}
			return false
		}))).
		Complete(r)
}
