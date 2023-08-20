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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var authcenterlog = logf.Log.WithName("authcenter-resource")

func (r *AuthCenter) SetupWebhookWithManager(mgr ctrl.Manager) error {
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
	if r.Spec.Username == "" {
		r.Spec.Username = "auth-" + r.Spec.Uid
	}
	if r.Status.Status == "" {
		r.Status.Status = StatusTypePending
	}
	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-auth-jiayi-com-v1-authcenter,mutating=false,failurePolicy=fail,sideEffects=None,groups=auth.jiayi.com,resources=authcenters,verbs=create;update,versions=v1,name=vauthcenter.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &AuthCenter{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *AuthCenter) ValidateCreate() (admission.Warnings, error) {
	authcenterlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *AuthCenter) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	authcenterlog.Info("validate update", "name", r.Name)
	auth, success := old.(*AuthCenter)
	if success && auth.Spec.Uid == "" {
		return admission.Warnings{"uid is required"}, nil
	}
	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *AuthCenter) ValidateDelete() (admission.Warnings, error) {
	authcenterlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
