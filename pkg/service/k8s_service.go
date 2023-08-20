package service

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/gookit/goutil"
	authv1 "jiayi.com/auth-center/api/v1"
	"jiayi.com/auth-center/pkg/config"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type K8sService struct {
	ctx    context.Context
	client client.Client
	log    logr.Logger
}

func NewK8sService(ctx context.Context, log logr.Logger, r client.Client) *K8sService {
	return &K8sService{
		ctx:    ctx,
		client: r,
		log:    log,
	}
}
func (s *K8sService) ApplyContainers(auth *authv1.AuthCenter) ([]authv1.ContainerPermissionStatus, error) {
	rollbackOfDelete, err := s.deleteContainers(auth.Spec.Uid, auth.Spec.ContainerItems)
	if err != nil {
		s.log.V(1).Error(err, "K8sService.ApplyContainers delete containers error")
		if config.GetContainerErrOfRollback() {
			s.rollbackContainersOfDelete(rollbackOfDelete)
		}
		return nil, err
	}
	rollbackOfApply, result, err := s.applyContainers(auth)
	if err != nil {
		s.log.V(1).Error(err, "K8sService.ApplyContainers apply containers error")
		if config.GetContainerErrOfRollback() {
			s.rollbackContainersOfApply(rollbackOfApply)
			s.rollbackContainersOfDelete(rollbackOfDelete)
		}
		return result, err
	}
	return result, nil
}

func (s *K8sService) deleteContainers(uid string, items []authv1.ContainerPermission) ([]rbacv1.RoleBinding, error) {
	nss := make([]string, 0)
	for _, item := range items {
		nss = append(nss, item.Namespace)
	}

	rbList := new(rbacv1.RoleBindingList)
	selector := fields.Set{
		"metadata.name": fmt.Sprintf(config.RbName, uid),
	}.AsSelector()
	listOpts := client.ListOptions{
		FieldSelector: selector,
		Namespace:     "",
	}
	err := s.client.List(s.ctx, rbList, &listOpts)
	if err != nil {
		s.log.V(1).Error(err, "list rolebinding error", "err", err)
		return nil, err
	}
	rollbackRbs := make([]rbacv1.RoleBinding, 0)
	for _, rb := range rbList.Items {
		if goutil.Contains(nss, rb.Namespace) {
			// 如果存在则不删除
			continue
		}
		err = s.client.Delete(s.ctx, &rb)
		if err != nil {
			s.log.V(1).Error(err, "delete rolebinding error", "err", err)
			return rollbackRbs, err
		}
		rollbackRbs = append(rollbackRbs, rb)
	}
	return rollbackRbs, nil
}

func (s *K8sService) rollbackContainersOfDelete(rbs []rbacv1.RoleBinding) {
	for _, rb := range rbs {
		rb.SetResourceVersion("")
		err := s.client.Create(s.ctx, &rb)
		if err != nil {
			s.log.V(1).Error(err, "K8sService.rollbackContainersOfDelete create rolebinding error", "err", err)
		}
	}
}

func (s *K8sService) applyContainers(auth *authv1.AuthCenter) ([]rbacv1.RoleBinding, []authv1.ContainerPermissionStatus, error) {

	uid := auth.Spec.Uid
	username := auth.Spec.Username
	items := auth.Spec.ContainerItems
	rollbackData := make([]rbacv1.RoleBinding, 0)
	result := make([]authv1.ContainerPermissionStatus, len(items))
	for index, item := range items {
		result[index] = authv1.ContainerPermissionStatus{
			ContainerPermission: item,
		}
		authName := config.ContainerReadOnly
		if item.AuthType == authv1.ReadOnly {
			authName = config.ContainerReadOnly
		} else if item.AuthType == authv1.Writable {
			authName = config.ContainerReadWrite
		} else if item.AuthType == authv1.Custom {
			authName = item.CustomRb
		}
		rb := new(rbacv1.RoleBinding)
		err := s.client.Get(s.ctx, types.NamespacedName{
			Namespace: item.Namespace,
			Name:      fmt.Sprintf(config.RbName, uid),
		}, rb)
		if err != nil {
			if client.IgnoreNotFound(err) != nil {
				s.log.V(1).Error(err, "K8sService.applyContainers get rolebinding error", "err", err)
				return rollbackData, result, err
			}
		} else {
			if authName == rb.RoleRef.Name {
				result[index].Status = true
				continue
			}
			err := s.client.Delete(s.ctx, rb)
			if err != nil {
				s.log.V(1).Error(err, "K8sService.applyContainers delete rolebinding error", "err", err)
				return rollbackData, result, err
			}
		}

		rbNew := &rbacv1.RoleBinding{
			TypeMeta: metav1.TypeMeta{Kind: "RoleBinding", APIVersion: "rbac.authorization.k8s.io/v1"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf(config.RbName, uid),
				Namespace: item.Namespace,
				Labels:    map[string]string{"uid": uid, "form": "authCenter"},
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: auth.APIVersion,
						Kind:       auth.Kind,
						Name:       auth.Name,
						UID:        auth.UID,
					},
				},
				Finalizers: []string{config.FinalizerNsAuthCenter},
			},
			Subjects: []rbacv1.Subject{
				{
					Kind: "User",
					Name: username,
				},
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     authName,
			},
		}
		err = s.client.Create(s.ctx, rbNew)
		if err != nil {
			s.log.V(1).Error(err, "K8sService.applyContainers create rolebinding error", "err", err)
			result[index].Status = false
			return rollbackData, result, err
		}
		result[index].Status = true
		if rb.CreationTimestamp.IsZero() {
			rb = rbNew
		}
		rollbackData = append(rollbackData, *rb)

	}
	return rollbackData, result, nil
}

func (s *K8sService) rollbackContainersOfApply(rbs []rbacv1.RoleBinding) {
	for _, rb := range rbs {
		err := s.client.Delete(s.ctx, &rb)
		if err != nil {
			s.log.V(1).Error(err, "K8sService.rollbackContainersOfApply delete rolebinding error", "err", err)
		}
		if !rb.CreationTimestamp.IsZero() {
			rb.SetResourceVersion("")
			err := s.client.Create(s.ctx, &rb)
			if err != nil {
				s.log.V(1).Error(err, "K8sService.rollbackContainersOfApply rollback rolebinding error", "err", err)
			}
		}
	}
}

func (s *K8sService) AddNsFinalizer(auth *authv1.AuthCenter) (int, error) {
	if len(auth.Spec.ContainerItems) == 0 {
		return 0, nil
	}
	operateNum := 0
	for _, item := range auth.Spec.ContainerItems {
		ns := new(corev1.Namespace)
		err := s.client.Get(s.ctx, types.NamespacedName{
			Name: item.Namespace,
		}, ns)
		if err != nil {
			s.log.V(1).Error(err, "K8sService.AddNsFinalizer get namespace error", "err", err)
			return operateNum, err
		}
		if !goutil.Contains(ns.ObjectMeta.Finalizers, config.FinalizerNsAuthCenter) {
			operateNum++
			ns.ObjectMeta.Finalizers = append(ns.ObjectMeta.Finalizers, config.FinalizerNsAuthCenter)
			err = s.client.Update(s.ctx, ns)
			if err != nil {
				s.log.V(1).Error(err, "K8sService.AddNsFinalizer update namespace error", "err", err)
				return operateNum, err
			}
		}
	}
	return operateNum, nil
}

func (s *K8sService) RemoveRoleBindingOfAuthCenterByNs(ns *corev1.Namespace) error {
	rblist := new(rbacv1.RoleBindingList)
	selector := labels.Set{
		"form": "authCenter",
	}.AsSelector()
	listOpts := client.ListOptions{
		LabelSelector: selector,
		Namespace:     ns.Name,
	}
	err := s.client.List(s.ctx, rblist, &listOpts)
	if err != nil {
		s.log.V(1).Error(err, "K8sService.RemoveRoleBindingOfAuthCenterByNs list rolebinding error", "err", err)
		return err
	}
	group := goutil.NewErrGroup(5)
	for _, rb := range rblist.Items {
		for _, reference := range rb.OwnerReferences {
			if reference.Kind == authv1.Kind {
				nsName := rb.Namespace
				group.Go(func() error {
					auth := new(authv1.AuthCenter)
					err := s.client.Get(s.ctx, types.NamespacedName{
						Name: reference.Name,
					}, auth)
					if err != nil {
						s.log.V(1).Error(err, "K8sService.RemoveRoleBindingOfAuthCenterByNs get authcenter error", "err", err)
						return err
					}
					auth.RemoveContainerItem(nsName)
					err = s.client.Update(s.ctx, auth)
					if err != nil {
						s.log.V(1).Error(err, "K8sService.RemoveRoleBindingOfAuthCenterByNs update authcenter error", "err", err)
						return err
					}
					return nil
				})
				break
			}
		}
	}
	err = group.Wait()
	return err
}

func (s *K8sService) RemoveRoleBindingOfAuthCenterByRoleBinding(rb *rbacv1.RoleBinding) error {
	for _, reference := range rb.OwnerReferences {
		if reference.Kind == authv1.Kind {
			auth := new(authv1.AuthCenter)
			err := s.client.Get(s.ctx, types.NamespacedName{
				Name: reference.Name,
			}, auth)
			if client.IgnoreNotFound(err) != nil {
				s.log.V(1).Error(err, "K8sService.RemoveRoleBindingOfAuthCenterByRoleBinding get authcenter error", "err", err)
				return err
			}
			if auth.GetContainerItemByNs(rb.Namespace) == nil {
				return nil
			}
			auth.RemoveContainerItem(rb.Namespace)
			err = s.client.Update(s.ctx, auth)
			if err != nil {
				s.log.V(1).Error(err, "K8sService.RemoveRoleBindingOfAuthCenterByRoleBinding update authcenter error", "err", err)
				return err
			}
			return nil
		}
	}
	return nil

}
