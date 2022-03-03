/*
Copyright 2022 ethanliu.

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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var operatordemolog = logf.Log.WithName("operatordemo-resource")

// 定义启动方法
func (r *OperatorDemo) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-ethanliuu-ethanliu-io-v1-operatordemo,mutating=true,failurePolicy=fail,sideEffects=None,groups=ethanliuu.ethanliu.io,resources=operatordemoes,verbs=create;update,versions=v1,name=moperatordemo.kb.io,admissionReviewVersions=v1

// 申明 defaulter
var _ webhook.Defaulter = &OperatorDemo{}

// 默认数据操作
// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *OperatorDemo) Default() {
	operatordemolog.Info("default", "name", r.Name)
	//定义如果Port为空时，给予定义默认值
	if r.Spec.Port == 0 {
		r.Spec.Port = 80
		operatordemolog.Info("set default value 80 for port")
	}
}

//+kubebuilder:webhook:path=/validate-ethanliuu-ethanliu-io-v1-operatordemo,mutating=false,failurePolicy=fail,sideEffects=None,groups=ethanliuu.ethanliu.io,resources=operatordemoes,verbs=create;update,versions=v1,name=voperatordemo.kb.io,admissionReviewVersions=v1

// 申明validator
var _ webhook.Validator = &OperatorDemo{}

// 创建 Validate 函数，创建时候要处理的逻辑
// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *OperatorDemo) ValidateCreate() error {
	operatordemolog.Info("validate create", "name", r.Name)
	return r.ValidateOperatorDemo()
}

// 更新 Validate 函数，更新时候要处理的逻辑
// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *OperatorDemo) ValidateUpdate(old runtime.Object) error {
	operatordemolog.Info("validate update", "name", r.Name)
	return r.ValidateOperatorDemo()
}

// 删除 Validate 函数，删除时候要处理的逻辑
// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *OperatorDemo) ValidateDelete() error {
	operatordemolog.Info("validate delete", "name", r.Name)
	return nil
}

// 判断规则是否通过，如果有不通过的则返回 apierrors
// 在 create 和 update 中执行该函数
func (r *OperatorDemo) ValidateOperatorDemo() error {
	var allErrs field.ErrorList
	if err := r.ValidatePort(); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(schema.GroupKind{
		Group: "ethanliuu.ethanliu.io",
		Kind:  "OperatorDemo",
	}, r.Name, allErrs)
}

// 定义准入规则，不允许 port 为 65535
func (r *OperatorDemo) ValidatePort() *field.Error {
	if r.Spec.Port == 65535 {
		return field.Invalid(field.NewPath("spec").Child("schedule"), r.Spec.Port, "port 不能为 65535")
	}
	return nil
}
