# operator-demo
operator学习记录

通过kubebuild 开发一个简单的crd，来快速创建极简的 deployment 和 service

**环境信息**
```
go1.17.1 darwin/amd64

KubeBuilderVersion:"3.3.0"

```
## 创建 operator
step1
```shell
# 创建项目目录
$ mkdir operator-demo 
$ cd operator-demo
#初始化项目目录
$ go mod init github.com/ethanliuuu/operator-demo   

$ kb init  --domain ethanliu.io --owner "ethanliu"

$ make manifests     
```
step2
```shell
# 创建API，两次确定
$ kb create api --group ethanliuu --version v1 --kind OperatorDemo  
Create Resource [y/n]
y
Create Controller [y/n]
y
```
step3
```go
# 修改 api/v1/operatordemo_types.go 

type OperatorDemoSpec struct {
	Image    string `json:"image"`
	Replicas int32  `json:"replicas"`
	// omitempty 非必选
	Port     int32  `json:"port,omitempty"`
}
```
step4
```go
# 修改 controllers/operatordemo_controller.go 中的策略
//+kubebuilder:rbac:groups=ethanliuu.ethanliu.io,resources=operatordemoes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ethanliuu.ethanliu.io,resources=operatordemoes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ethanliuu.ethanliu.io,resources=operatordemoes/finalizers,verbs=update
因为要对 deployment 和 service 操作，所以需要这两个资源的权限。
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services/status,verbs=get;update;patch
```
```go
# 添加主逻辑
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
```
step5
```shell
生成 RBAC 配置
$ make manifests
安装crd到集群里
$ make install
```
step6
```shell
#修改镜像地址
cat config/default/manager_auth_proxy_patch.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: controller-manager
  namespace: system
spec:
  template:
    spec:
      containers:
      - name: kube-rbac-proxy
        image: registry.cn-beijing.aliyuncs.com/ethanliuu/kube-rbac-proxy:v0.8.0
        args:
        - "--secure-listen-address=0.0.0.0:8443"
        - "--upstream=http://127.0.0.1:8080/"
        - "--logtostderr=true"
        - "--v=0"
        ports:
        - containerPort: 8443
          protocol: TCP
          name: https
        resources:
          limits:
            cpu: 500m
            memory: 128Mi
          requests:
            cpu: 5m
            memory: 64Mi
      - name: manager
        args:
        - "--health-probe-bind-address=:8081"
        - "--metrics-bind-address=127.0.0.1:8080"
        - "--leader-elect"
```
step7
```shell
#编译、镜像构建
$ expor IMG=registry.cn-beijing.aliyuncs.com/ethanliuu/operator-demo:1.0.0
$ make docker-build doker-push

#将控制器部署到集群里
$make deploy


$ k get pod -n operator-demo-system 
NAME                                                READY   STATUS    RESTARTS   AGE
operator-demo-controller-manager-5d5447cfcb-9jhbq   2/2     Running   0          1m

$ k get crd operatordemoes.ethanliuu.ethanliu.io 
NAME                                   CREATED AT
operatordemoes.ethanliuu.ethanliu.io   2022-03-02T02:10:08Z
```
step7
```shell
# 修改 yaml ，测试 operator
$ cat config/samples/ethanliuu_v1_operatordemo.yaml
apiVersion: ethanliuu.ethanliu.io/v1
kind: OperatorDemo
metadata:
  name: operatordemo-sample
spec:
  image: nginx:latest
  replicas: 2
  port: 80
  
$ k apply -f config/samples/ethanliuu_v1_operatordemo.yaml

$ k get pod,svc
NAME                                                    READY   STATUS    RESTARTS   AGE
pod/operatordemo-sample-78bffc997-5jnjc                 1/1     Running   0          31m
pod/operatordemo-sample-78bffc997-d8sd2                 1/1     Running   0          31m

NAME                                                       TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
service/operatordemo-sample                                ClusterIP   10.110.32.102    <none>        80/TCP     31m
```
## 创建webhook
webhook是什么？

我理解的 webhook 就是在 controller 做处理之前，由 webhook 做一些字段的合法性校验，字段的默认值填充，确保一份真实、可用的资源清单交给 controller。

部署 cert-manager
```shell
$ k apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.7.1/cert-manager.yaml
```
添加 webhook
```shell
$ kb create webhook --group ethanliuu --version v1 --kind OperatorDemo --defaulting --programmatic-validation
```
修改webhook逻辑
```go
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
```
**修改config**

**config/default/kustomization.yaml**

因为需要使用 webhook，所以需要打开 webhook 和 cert-manager的注释
```yaml
# Adds namespace to all resources.
namespace: operator-demo-system

# Value of this field is prepended to the
# names of all resources, e.g. a deployment named
# "wordpress" becomes "alices-wordpress".
# Note that it should also match with the prefix (text before '-') of the namespace
# field above.
namePrefix: operator-demo-

# Labels to add to all resources and selectors.
#commonLabels:
#  someName: someValue

bases:
- ../crd
- ../rbac
- ../manager
# [WEBHOOK] To enable webhook, uncomment all the sections with [WEBHOOK] prefix including the one in
# crd/kustomization.yaml
- ../webhook
# [CERTMANAGER] To enable cert-manager, uncomment all sections with 'CERTMANAGER'. 'WEBHOOK' components are required.
- ../certmanager
# [PROMETHEUS] To enable prometheus monitor, uncomment all sections with 'PROMETHEUS'.
#- ../prometheus

patchesStrategicMerge:
# Protect the /metrics endpoint by putting it behind auth.
# If you want your controller-manager to expose the /metrics
# endpoint w/o any authn/z, please comment the following line.
- manager_auth_proxy_patch.yaml

# Mount the controller config file for loading manager configurations
# through a ComponentConfig type
#- manager_config_patch.yaml

# [WEBHOOK] To enable webhook, uncomment all the sections with [WEBHOOK] prefix including the one in
# crd/kustomization.yaml
- manager_webhook_patch.yaml

# [CERTMANAGER] To enable cert-manager, uncomment all sections with 'CERTMANAGER'.
# Uncomment 'CERTMANAGER' sections in crd/kustomization.yaml to enable the CA injection in the admission webhooks.
# 'CERTMANAGER' needs to be enabled to use ca injection
- webhookcainjection_patch.yaml

# the following config is for teaching kustomize how to do var substitution
vars:
- name: CERTIFICATE_NAMESPACE # namespace of the certificate CR
  objref:
    kind: Certificate
    group: cert-manager.io
    version: v1
    name: serving-cert # this name should match the one in certificate.yaml
  fieldref:
    fieldpath: metadata.namespace
- name: CERTIFICATE_NAME
  objref:
    kind: Certificate
    group: cert-manager.io
    version: v1
    name: serving-cert # this name should match the one in certificate.yaml
- name: SERVICE_NAMESPACE # namespace of the service
  objref:
    kind: Service
    version: v1
    name: webhook-service
  fieldref:
    fieldpath: metadata.namespace
- name: SERVICE_NAME
  objref:
    kind: Service
    version: v1
    name: webhook-service

```

config/crd/kustomization.yaml
```yaml
# This kustomization.yaml is not intended to be run by itself,
# since it depends on service name and namespace that are out of this kustomize package.
# It should be run by config/default
resources:
- bases/ethanliuu.ethanliu.io_operatordemoes.yaml
#+kubebuilder:scaffold:crdkustomizeresource

patchesStrategicMerge:
# [WEBHOOK] To enable webhook, uncomment all the sections with [WEBHOOK] prefix.
# patches here are for enabling the conversion webhook for each CRD
- patches/webhook_in_operatordemoes.yaml
#+kubebuilder:scaffold:crdkustomizewebhookpatch

# [CERTMANAGER] To enable cert-manager, uncomment all the sections with [CERTMANAGER] prefix.
# patches here are for enabling the CA injection for each CRD
- patches/cainjection_in_operatordemoes.yaml
#+kubebuilder:scaffold:crdkustomizecainjectionpatch

# the following config is for teaching kustomize how to do kustomization for CRDs.
configurations:
- kustomizeconfig.yaml
```
**编译**
```shell
$ export IMG=registry.cn-beijing.aliyuncs.com/ethanliuu/operator-demo-webhook:1.0.2
$ make docker-build docker-push
```
**部署**
```shell
$ make deploy
```
测试 Default
```shell
1.6462875937064295e+09	INFO	operatordemo-resource	default	{"name": "operatordemo-sample"}
1.646287593706446e+09	INFO	operatordemo-resource	set default value 80 for port
```
验证准入
```shell
$ k apply -f config/samples/ethanliuu_v1_operatordemo.yaml 
The OperatorDemo "operatordemo-sample" is invalid: spec.schedule: Invalid value: 65535: port 不能为 65535

```
[参考好朋友的项目](https://github.com/barrettzjh/test-operator)
