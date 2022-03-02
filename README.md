# operator-demo
operator学习记录

通过kubebuild 开发一个简单的crd，来快速创建极简的 deployment 和 service

**环境信息**
```
go1.17.1 darwin/amd64

KubeBuilderVersion:"3.3.0"

```
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
	Port     int32  `json:"port"`
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
[参考好朋友的项目](https://github.com/barrettzjh/test-operator)
