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

package controllers

import (
	"context"
	operatorv1 "github.com/ethanliuuu/operator-demo/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// OperatorDemoReconciler reconciles a OperatorDemo object
type OperatorDemoReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ethanliuu.ethanliu.io,resources=operatordemoes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ethanliuu.ethanliu.io,resources=operatordemoes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ethanliuu.ethanliu.io,resources=operatordemoes/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the OperatorDemo object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *OperatorDemoReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	// 判断当前 Reconcile 应该做什么操作
	var operatorDemo operatorv1.OperatorDemo
	//判断该 crd 还存不存在，如果已经存在，清理资源
	operatorDemo, err := r.isCrdExist(ctx, req)
	if err != nil {
		r.clear(ctx, req)
		return ctrl.Result{}, nil
	}

	//存在则更新，不存在则创建
	if dep, err := r.isDeployExist(ctx, req); err != nil {
		log.Log.Info("deployment is not exists,create deployment")
		if err := r.createDeployment(ctx, operatorDemo); err != nil {
			log.Log.Info("create deployment failed,err:", err.Error())
		}
	} else {
		log.Log.Info("deployment already exists,update deployment.")
		r.updateDeployment(ctx, operatorDemo, dep)
	}
	if service, err := r.isServiceExist(ctx, req); err != nil {
		log.Log.Info("service is not exists,create service")
		if err := r.createService(ctx, operatorDemo); err != nil {
			log.Log.Info("create service failed,err:", "err", err.Error())
		}
	} else {
		log.Log.Info("service already exists,update service")
		r.updateService(ctx, operatorDemo, service)
	}
	return ctrl.Result{}, nil
}

func (r *OperatorDemoReconciler) isCrdExist(ctx context.Context, req ctrl.Request) (operatorv1.OperatorDemo, error) {
	var crd operatorv1.OperatorDemo
	err := r.Get(ctx, req.NamespacedName, &crd)
	return crd, err
}

//清理资源
func (r *OperatorDemoReconciler) clear(ctx context.Context, req ctrl.Request) {
	if deployment, err := r.isDeployExist(ctx, req); err == nil {
		if err := r.Client.Delete(ctx, &deployment); err != nil {
			log.Log.Info("delete deployment failed，name=" + deployment.Name)
		}
	}

	if service, err := r.isServiceExist(ctx, req); err == nil {
		if err := r.Client.Delete(ctx, &service); err != nil {
			log.Log.Info("delete service failed, name=" + service.Name)

		}
	}
}

//判断deployment是否存在
func (r *OperatorDemoReconciler) isDeployExist(ctx context.Context, req ctrl.Request) (appsv1.Deployment, error) {
	var dep appsv1.Deployment
	err := r.Get(ctx, req.NamespacedName, &dep)
	return dep, err
}

// 判断service是否存在
func (r *OperatorDemoReconciler) isServiceExist(ctx context.Context, req ctrl.Request) (corev1.Service, error) {
	var service corev1.Service
	err := r.Get(ctx, req.NamespacedName, &service)
	return service, err
}

func (r *OperatorDemoReconciler) createDeployment(ctx context.Context, crd operatorv1.OperatorDemo) error {
	log.Log.Info("Create Deployment")
	if err := r.Client.Create(ctx, getDeployment(crd)); err != nil {
		return err
	}
	return nil
}

func (r *OperatorDemoReconciler) createService(ctx context.Context, crd operatorv1.OperatorDemo) error {
	log.Log.Info("Create Service")
	if err := r.Client.Create(ctx, getService(crd)); err != nil {
		return err
	}
	return nil
}

func (r *OperatorDemoReconciler) updateDeployment(ctx context.Context, crd operatorv1.OperatorDemo, deployment appsv1.Deployment) error {
	deployment.Spec.Replicas = &crd.Spec.Replicas
	deployment.Spec.Template.Spec.Containers[0].Image = crd.Spec.Image
	if err := r.Client.Update(ctx, &deployment); err != nil {
		return err
	}
	return nil
}

func (r *OperatorDemoReconciler) updateService(ctx context.Context, crd operatorv1.OperatorDemo, service corev1.Service) error {
	log.Log.Info("Update Service")
	service.Spec.Ports[0].Port = crd.Spec.Port
	//service.Spec.Ports[0].NodePort = crd.Spec.NodePort
	if err := r.Client.Update(ctx, &service); err != nil {
		return err
	}
	return nil
}

func getDeployment(s operatorv1.OperatorDemo) *appsv1.Deployment {
	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name,
			Namespace: s.Namespace,
			Labels:    map[string]string{"app": s.Name},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(s.Spec.Replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": s.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      s.Name,
					Namespace: s.Namespace,
					Labels:    map[string]string{"app": s.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  s.Name,
							Image: s.Spec.Image,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: s.Spec.Port,
								},
							},
						},
					},
				},
			},
		},
	}
	return dep
}

func getService(s operatorv1.OperatorDemo) *corev1.Service {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name,
			Namespace: s.Namespace,
			Labels:    map[string]string{"app": s.Name},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       s.Name,
					Protocol:   corev1.ProtocolTCP,
					Port:       s.Spec.Port,
					TargetPort: intstr.IntOrString{IntVal: s.Spec.Port},
					//NodePort:   s.Spec.NodePort,
				},
			},
			//Type:     "NodePort",
			Selector: map[string]string{"app": s.Name},
		},
	}
	return service
}

func int32Ptr(i int32) *int32 {
	return &i
}

// SetupWithManager sets up the controller with the Manager.
func (r *OperatorDemoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1.OperatorDemo{}).
		Complete(r)
}
