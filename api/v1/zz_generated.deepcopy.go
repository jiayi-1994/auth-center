//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AuthCenter) DeepCopyInto(out *AuthCenter) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AuthCenter.
func (in *AuthCenter) DeepCopy() *AuthCenter {
	if in == nil {
		return nil
	}
	out := new(AuthCenter)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AuthCenter) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AuthCenterList) DeepCopyInto(out *AuthCenterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AuthCenter, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AuthCenterList.
func (in *AuthCenterList) DeepCopy() *AuthCenterList {
	if in == nil {
		return nil
	}
	out := new(AuthCenterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AuthCenterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AuthCenterSpec) DeepCopyInto(out *AuthCenterSpec) {
	*out = *in
	out.Harbor = in.Harbor
	if in.ContainerItems != nil {
		in, out := &in.ContainerItems, &out.ContainerItems
		*out = make([]ContainerPermission, len(*in))
		copy(*out, *in)
	}
	if in.HarborItems != nil {
		in, out := &in.HarborItems, &out.HarborItems
		*out = make([]HarborPermission, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AuthCenterSpec.
func (in *AuthCenterSpec) DeepCopy() *AuthCenterSpec {
	if in == nil {
		return nil
	}
	out := new(AuthCenterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AuthCenterStatus) DeepCopyInto(out *AuthCenterStatus) {
	*out = *in
	if in.ContainerItems != nil {
		in, out := &in.ContainerItems, &out.ContainerItems
		*out = make([]ContainerPermissionStatus, len(*in))
		copy(*out, *in)
	}
	if in.HarborItems != nil {
		in, out := &in.HarborItems, &out.HarborItems
		*out = make([]HarborPermissionStatus, len(*in))
		copy(*out, *in)
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]AuthCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AuthCenterStatus.
func (in *AuthCenterStatus) DeepCopy() *AuthCenterStatus {
	if in == nil {
		return nil
	}
	out := new(AuthCenterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AuthCondition) DeepCopyInto(out *AuthCondition) {
	*out = *in
	in.LastProbeTime.DeepCopyInto(&out.LastProbeTime)
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AuthCondition.
func (in *AuthCondition) DeepCopy() *AuthCondition {
	if in == nil {
		return nil
	}
	out := new(AuthCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ContainerPermission) DeepCopyInto(out *ContainerPermission) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ContainerPermission.
func (in *ContainerPermission) DeepCopy() *ContainerPermission {
	if in == nil {
		return nil
	}
	out := new(ContainerPermission)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ContainerPermissionStatus) DeepCopyInto(out *ContainerPermissionStatus) {
	*out = *in
	out.ContainerPermission = in.ContainerPermission
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ContainerPermissionStatus.
func (in *ContainerPermissionStatus) DeepCopy() *ContainerPermissionStatus {
	if in == nil {
		return nil
	}
	out := new(ContainerPermissionStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HarborInfo) DeepCopyInto(out *HarborInfo) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HarborInfo.
func (in *HarborInfo) DeepCopy() *HarborInfo {
	if in == nil {
		return nil
	}
	out := new(HarborInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HarborPermission) DeepCopyInto(out *HarborPermission) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HarborPermission.
func (in *HarborPermission) DeepCopy() *HarborPermission {
	if in == nil {
		return nil
	}
	out := new(HarborPermission)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HarborPermissionStatus) DeepCopyInto(out *HarborPermissionStatus) {
	*out = *in
	out.HarborPermission = in.HarborPermission
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HarborPermissionStatus.
func (in *HarborPermissionStatus) DeepCopy() *HarborPermissionStatus {
	if in == nil {
		return nil
	}
	out := new(HarborPermissionStatus)
	in.DeepCopyInto(out)
	return out
}
