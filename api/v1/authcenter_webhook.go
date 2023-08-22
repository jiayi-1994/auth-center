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
	"context"
	"errors"

	"github.com/gookit/goutil"
	harborClient "jiayi.com/auth-center/pkg/client"
	authUtil "jiayi.com/auth-center/pkg/util"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var authcenterlog = logf.Log.WithName("authcenter-resource")
var cli client.Client
var hbCli *harborClient.HarborClient

func (r *AuthCenter) SetupWebhookWithManager(mgr ctrl.Manager) error {
	cli = mgr.GetClient()
	hbCli = harborClient.GetHarborClient(context.Background())
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-auth-jiayi-com-v1-authcenter,mutating=true,failurePolicy=fail,sideEffects=None,groups=auth.jiayi.com,resources=authcenters,verbs=create;update,versions=v1,name=mauthcenter.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &AuthCenter{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *AuthCenter) Default() {
	authcenterlog.Info("default", "name", r.Name)

	if r.Status.Status == "" {
		r.Status.Status = StatusTypePending
	}
	projects, err := hbCli.ListAllProjects(nil)
	if err != nil {
		authcenterlog.Error(err, "list all projects error")
		return
	}
	pjs := make([]string, 0)
	for _, project := range projects {
		pjs = append(pjs, goutil.String(project.ProjectID))
	}
	// hb项目不存在 则remove
	items := make([]HarborPermission, 0)
	for _, item := range r.Spec.HarborItems {
		if goutil.Contains(pjs, item.ProjectID) {
			items = append(items, item)
		}
	}
	r.Spec.HarborItems = items

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-auth-jiayi-com-v1-authcenter,mutating=false,failurePolicy=fail,sideEffects=None,groups=auth.jiayi.com,resources=authcenters,verbs=create;update,versions=v1,name=vauthcenter.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &AuthCenter{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *AuthCenter) ValidateCreate() (admission.Warnings, error) {
	authcenterlog.Info("validate create", "name", r.Name)
	return r.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *AuthCenter) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	authcenterlog.Info("validate update", "name", r.Name)
	auth, success := old.(*AuthCenter)
	if success && auth.Spec.Uid == "" || r.Spec.Uid == "" {
		return admission.Warnings{"uid is required"}, errors.New("uid is required")
	}
	return r.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *AuthCenter) ValidateDelete() (admission.Warnings, error) {
	authcenterlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

func (r *AuthCenter) Validate() (admission.Warnings, error) {
	if r.Spec.Uid == "" {
		return admission.Warnings{"uid is required"}, errors.New("uid is required")
	}
	if r.Spec.Username == "" {
		return admission.Warnings{"username is required"}, errors.New("username is required")
	}
	if r.Spec.Harbor.Password != "" && !authUtil.ValidatePassword(r.Spec.Harbor.Password) {
		return admission.Warnings{"The password must be between 8 and 20 long and contain at least one uppercase character, one lowercase character, and one number"}, errors.New("the password must be between 8 and 20 long and contain at least one uppercase character, one lowercase character, and one number")
	}

	if r.Spec.Harbor.EncryptPwd != "" {
		if _, err := authUtil.DecryptByAes(r.Spec.Harbor.EncryptPwd); err != nil {
			return admission.Warnings{"decrypt password error"}, errors.New("decrypt password error")
		}
	}

	if len(r.Spec.HarborItems) > 0 && (r.Spec.Harbor.HarborUid == "" && r.Spec.Harbor.Name == "") {
		return admission.Warnings{"Spec.HarborItems is not empty, you must config spec.harbor user info"}, errors.New("spec.harborItems is not empty, you must config spec.harbor user info")
	}
	return nil, nil
}
