/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	authv1 "jiayi.com/auth-center/api/v1"
	"jiayi.com/auth-center/pkg/config"
	"jiayi.com/auth-center/pkg/service"
	"jiayi.com/auth-center/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// AuthCenterReconciler reconciles a AuthCenter object
type AuthCenterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=auth.jiayi.com,resources=authcenters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=auth.jiayi.com,resources=authcenters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=auth.jiayi.com,resources=authcenters/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=*,verbs=*

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AuthCenter object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *AuthCenterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("traceID", time.Now().UnixNano())

	authCenter := new(authv1.AuthCenter)
	err := r.Get(ctx, req.NamespacedName, authCenter)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			logger.V(1).Error(err, "Get authCenter error, try later or ignore when not found")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		logger.Info("delete authCenter end")
	}

	logger.Info("Start processing ", "authCenter", authCenter)
	harborService := service.NewHarborService(ctx, logger, &r.Client)
	// 如果进入删除状态 Finalizers 要先进行删除 第三方相关权限内容，如harbor， k8s资源根据owner 会自动删除
	if !authCenter.DeletionTimestamp.IsZero() {
		if len(authCenter.ObjectMeta.Finalizers) == 0 {
			return ctrl.Result{}, nil
		}
		if authCenter.Status.Status != authv1.StatusTerminating {
			authCenter.Status.Status = authv1.StatusTerminating
			err = r.Status().Update(ctx, authCenter)
			if err != nil {
				return ctrl.Result{}, err
			}
		}

		// TODO remove Harbor
		if config.AllCfg.Harbor.Enable {
			err := harborService.DeleteUser(authCenter)
			if err != nil {
				logger.Error(err, "delete harbor user error")
				return ctrl.Result{}, err
			}
			authCenter.ObjectMeta.Finalizers = util.DeleteSliceElement(authCenter.ObjectMeta.Finalizers, config.FinalizerHarbor)
		}

		err = r.Update(ctx, authCenter)
	}

	if r.IsReturn(authCenter) || authCenter.CreationTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	// 前置处理不通过则进入pending状态 如初始化ns finalizer ， 初始化harbor用户
	k8sService := service.NewK8sService(ctx, logger, r.Client)
	operateNum, err := k8sService.AddNsFinalizer(authCenter)
	if err != nil {
		logger.Info("add ns finalizer error")
		r.Record(ctx, authCenter, authv1.InitNamespaceFinalizer, err, authv1.StatusTypePending, logger)
		return ctrl.Result{}, err
	} else if operateNum > 0 || authCenter.GetConditionIndexByType(authv1.InitNamespaceFinalizer) == -1 {
		r.Record(ctx, authCenter, authv1.InitNamespaceFinalizer, err, "", logger)
		return ctrl.Result{}, nil
	}

	if config.AllCfg.Harbor.Enable {
		if isReturn, err := r.AddFinalizers(ctx, authCenter, config.FinalizerHarbor); isReturn || err != nil {
			return ctrl.Result{}, err
		}

		isReturn, err := harborService.InitHarborUser(authCenter)
		if err != nil {
			logger.Info("init harbor user error")
			r.Record(ctx, authCenter, authv1.InitHarborUser, err, authv1.StatusTypePending, logger)
			return ctrl.Result{}, err
		} else {
			if isReturn {
				return ctrl.Result{}, r.Update(ctx, authCenter)
			} else if authCenter.GetConditionIndexByType(authv1.InitHarborUser) == -1 {
				r.Record(ctx, authCenter, authv1.InitHarborUser, err, "", logger)
			}
		}
	}

	//处理k8s相关权限
	logger.Info("dealing with container permissions start")
	isContainerUpdate := authCenter.IsUpdateContainer()
	result, err := k8sService.ApplyContainers(authCenter)
	authCenter.Status.ContainerItems = result
	logger.Info("dealing with container permissions end")
	if isContainerUpdate {
		logger.Info("dealing with container permissions: has update")
		r.Record(ctx, authCenter, authv1.ContainersRbacReady, err, "", logger)
		return ctrl.Result{}, err
	}
	//处理harbor相关权限
	if config.AllCfg.Harbor.Enable {
		logger.Info("dealing with harbor permissions start")
		isHarborUpdate := authCenter.IsUpdateHarbor()
		auths, err := harborService.ApplyHarborAuth(authCenter)
		authCenter.Status.HarborItems = auths
		if isHarborUpdate {
			logger.Info("dealing with harbor permissions: has update")
			r.Record(ctx, authCenter, authv1.HarborReady, err, "", logger)
			return ctrl.Result{}, err
		}
	}

	// end
	authCenter.Status.Status = authv1.StatusTypeSuccess
	err = r.Status().Update(ctx, authCenter)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// AddFinalizers true need to return
func (r *AuthCenterReconciler) AddFinalizers(ctx context.Context, authCenter *authv1.AuthCenter, key string) (bool, error) {
	if util.ContainsString(authCenter.ObjectMeta.Finalizers, key) {
		return false, nil
	}
	authCenter.ObjectMeta.Finalizers = append(authCenter.ObjectMeta.Finalizers, key)
	err := r.Update(ctx, authCenter)
	if err != nil {
		return true, err
	}
	return true, nil
}

func (r *AuthCenterReconciler) Record(ctx context.Context, authCenter *authv1.AuthCenter, conditionType authv1.ConditionType, err error, errStatusType authv1.StatusType, logger logr.Logger) {
	authCenter.Status.Status = authv1.StatusTypeRunning
	nowTime := metav1.NewTime(time.Now())
	message := ""
	if err != nil {
		message = err.Error()
		authCenter.Status.Status = errStatusType
		if errStatusType == "" {
			authCenter.Status.Status = authv1.StatusTypeFailed
		}
	}
	conditionIndex := authCenter.GetConditionIndexByType(conditionType)
	if conditionIndex == -1 {
		condition := authv1.AuthCondition{
			Status:             err == nil,
			Type:               conditionType,
			LastProbeTime:      nowTime,
			LastTransitionTime: nowTime,
			Message:            message,
		}
		authCenter.Status.Conditions = append(authCenter.Status.Conditions, condition)
	} else {
		authCenter.Status.Conditions[conditionIndex].Status = err == nil
		authCenter.Status.Conditions[conditionIndex].LastProbeTime = nowTime
		authCenter.Status.Conditions[conditionIndex].LastTransitionTime = nowTime
		authCenter.Status.Conditions[conditionIndex].Message = message
	}
	err = r.Status().Update(ctx, authCenter)
	if err != nil {
		logger.Error(err, "Update authCenter error")
	}
}

func (r *AuthCenterReconciler) IsReturn(authCenter *authv1.AuthCenter) bool {
	return !authCenter.IsUpdateContainer() && !authCenter.IsUpdateHarbor() && (authCenter.Status.Status == authv1.StatusTypeSuccess || authCenter.Status.Status == authv1.StatusTypeFailed)
}

// SetupWithManager sets up the controller with the Manager.
func (r *AuthCenterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&authv1.AuthCenter{}).
		Complete(r)
}
