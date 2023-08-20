package controller

import (
	"context"
	"time"

	"jiayi.com/auth-center/pkg/config"
	"jiayi.com/auth-center/pkg/service"

	"jiayi.com/auth-center/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type NamespaceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *NamespaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("traceID", time.Now().UnixNano())
	logger.Info("NamespaceReconciler.Reconcile start")
	ns := new(corev1.Namespace)
	err := r.Get(ctx, req.NamespacedName, ns)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			logger.Error(err, "NamespaceReconciler.Reconcile get namespace error")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		return ctrl.Result{}, nil
	}
	if !ns.DeletionTimestamp.IsZero() {
		err := service.NewK8sService(ctx, logger, r.Client).RemoveRoleBindingOfAuthCenterByNs(ns)
		if err != nil {
			logger.Error(err, "NamespaceReconciler.Reconcile remove rolebinding error")
			return ctrl.Result{}, err
		}
		ns.ObjectMeta.Finalizers = util.RemoveString(ns.ObjectMeta.Finalizers, config.FinalizerNsAuthCenter)
		err = r.Update(ctx, ns)
		return ctrl.Result{}, err
	}
	logger.Info("NamespaceReconciler.Reconcile end")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}, builder.WithPredicates(predicate.NewPredicateFuncs(func(o client.Object) bool {
			if util.ContainsString(o.GetFinalizers(), config.FinalizerNsAuthCenter) {
				return true
			}
			return false
		}))).
		Complete(r)
}
