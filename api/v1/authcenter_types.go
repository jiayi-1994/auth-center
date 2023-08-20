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

package v1

import (
	"jiayi.com/auth-center/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StatusType string // 状态类型
type AuthType string
type ConditionType string

const (
	InitNamespaceFinalizer ConditionType = "InitNamespaceFinalizer"
	ContainersRbacReady    ConditionType = "ContainersRbacReady"
	HarborReady            ConditionType = "HarborReady"

	StatusTypePending StatusType = "Pending"
	StatusTypeSuccess StatusType = "Success"
	StatusTypeFailed  StatusType = "Failed"
	StatusTypeRunning StatusType = "Running"
	StatusTypeUnknown StatusType = "Unknown"

	ReadOnly AuthType = "readonly" //只读
	Writable AuthType = "writable" //读写
	Custom   AuthType = "custom"   //自定义 自己配置role权限

	Kind = "AuthCenter"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AuthCenterSpec defines the desired state of AuthCenter
type AuthCenterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// 用户唯一主键id
	Uid            string                `json:"uid"`                      // 用户唯一主键id
	Username       string                `json:"username"`                 // 用户名 也是 绑定roleBinging中User的名称 也是用于生成kubeconfig字段
	HarborUid      string                `json:"harborUid,omitempty"`      // 用户ID关联的harbor用户ID
	ContainerItems []ContainerPermission `json:"containerItems,omitempty"` // 容器权限列表
	HarborItems    []HarborPermission    `json:"harborItems,omitempty"`    // harbor权限列表
}
type ContainerPermission struct {
	AuthType  AuthType `json:"authType,omitempty"`  // 是否只读
	CustomRb  string   `json:"customRb,omitempty"`  // 自定义配置ClusterRole
	Namespace string   `json:"namespace,omitempty"` // 命名空间
}

type HarborPermission struct {
	ProjectID int64 `json:"projectId,omitempty"` // 项目ID
	RoleID    int64 `json:"roleId,omitempty"`    // 角色ID 1 个用于项目管理员，2 个用于开发人员，3 个用于来宾，4 个用于维护者
}

// AuthCenterStatus defines the observed state of AuthCenter
type AuthCenterStatus struct {
	Status StatusType `json:"status,omitempty"`
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// A list of current conditions of the resource
	ContainerItems []ContainerPermissionStatus `json:"containerItems,omitempty"` // 容器权限列表
	HarborItems    []HarborPermissionStatus    `json:"harborItems,omitempty"`    // harbor权限列表
	Conditions     []AuthCondition             `json:"conditions,omitempty" optional:"true"`
}
type ContainerPermissionStatus struct {
	ContainerPermission `json:",inline"`
	Status              bool `json:"status,omitempty"` // 是否成功
}
type HarborPermissionStatus struct {
	HarborPermission `json:",inline"`
	Status           bool `json:"status,omitempty"` // 是否成功
}

type AuthCondition struct {
	Status bool          `json:"status,omitempty"`
	Type   ConditionType `json:"type,omitempty"`
	// Last time we probed the condition.
	// +optional
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty" protobuf:"bytes,3,opt,name=lastProbeTime"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,4,opt,name=lastTransitionTime"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster,shortName=ac,categories=auth
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Uid",type="string",JSONPath=".spec.uid",description="用户ID 主键"
//+kubebuilder:printcolumn:name="Username",type="string",JSONPath=".spec.username",description="用户名 也是 绑定roleBinging中User的名称 也是用于生成kubeconfig字段"
//+kubebuilder:printcolumn:name="HarborUid",type="string",JSONPath=".spec.harborUid",description="harbor 绑定的id"
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status",description="状态"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Age for this resource"

// AuthCenter is the Schema for the authcenters API
type AuthCenter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AuthCenterSpec   `json:"spec,omitempty"`
	Status AuthCenterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AuthCenterList contains a list of AuthCenter
type AuthCenterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AuthCenter `json:"items"`
}

func (auth *AuthCenter) GetConditionIndexByType(t ConditionType) int {
	for index, condition := range auth.Status.Conditions {
		if condition.Type == t {
			return index
		}
	}
	return -1
}
func (auth *AuthCenter) IsUpdateContainer() bool {
	if auth.GetConditionIndexByType(ContainersRbacReady) == -1 {
		return true
	}
	if len(auth.Spec.ContainerItems) != len(auth.Status.ContainerItems) {
		return true
	}
	for _, item := range auth.Spec.ContainerItems {
		flag := false
		for _, status := range auth.Status.ContainerItems {
			if status.Status && item.Namespace == status.Namespace && item.AuthType == status.AuthType {
				flag = true
				break
			}
		}
		if !flag {
			return true
		}
	}

	return false
}
func (auth *AuthCenter) IsUpdateHarbor() bool {
	if !config.AllCfg.Harbor.Enable {
		return false
	}
	if len(auth.Spec.HarborItems) == 0 && len(auth.Status.HarborItems) == 0 {
		return false
	}
	if auth.GetConditionIndexByType(HarborReady) == -1 {
		return true
	}
	if len(auth.Spec.HarborItems) != len(auth.Status.HarborItems) {
		return false
	}
	for _, item := range auth.Spec.HarborItems {
		flag := false
		for _, status := range auth.Status.HarborItems {
			if status.Status && item.ProjectID == status.ProjectID && item.RoleID == status.RoleID {
				flag = true
				break
			}
		}
		if !flag {
			return true
		}
	}
	return false
}

func (auth *AuthCenter) RemoveContainerItem(ns string) {
	for index, item := range auth.Spec.ContainerItems {
		if item.Namespace == ns {
			auth.Spec.ContainerItems = append(auth.Spec.ContainerItems[:index], auth.Spec.ContainerItems[index+1:]...)
			break
		}
	}
}
func (auth *AuthCenter) GetContainerItemByNs(ns string) *ContainerPermission {
	for _, item := range auth.Spec.ContainerItems {
		if item.Namespace == ns {
			return &item
		}
	}
	return nil
}

func init() {
	SchemeBuilder.Register(&AuthCenter{}, &AuthCenterList{})
}
