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

package webhook

import (
	"context"
	"github.com/go-logr/logr"
	"gitlab.wellcloud.cc/cloud/dictator/registry"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps,resources=deployments/finalizers,verbs=update

//+kubebuilder:webhook:path=/mutate-apps-v1-deployment,mutating=true,failurePolicy=fail,sideEffects=None,groups=apps,resources=deployments,verbs=create;update,versions=v1,name=mdeployment.kb.io,admissionReviewVersions=v1
//+kubebuilder:webhook:path=/validate-apps-v1-deployment,mutating=false,failurePolicy=fail,sideEffects=None,groups=apps,resources=deployments,verbs=create;update,versions=v1,name=vdeployment.kb.io,admissionReviewVersions=v1

type DeploymentWebhook struct {
	client client.Client
	logger logr.Logger
}

func SetupDeploymentWebhookWithManager(mgr ctrl.Manager) error {
	hook := &DeploymentWebhook{
		client: mgr.GetClient(),
		logger: logf.Log.WithName("[webhook.deployment]"),
	}
	return ctrl.NewWebhookManagedBy(mgr).
		For(&appsv1.Deployment{}).
		WithDefaulter(hook).
		WithValidator(hook).
		Complete()
}

const (
	K8sAnnotationDependence = ".wkm.welljoint.com/dependence" // 依赖约束
)

func (w *DeploymentWebhook) Default(ctx context.Context, obj runtime.Object) error {
	return UseDefault(obj, w.logger)
}

func (w *DeploymentWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	w.logger.Info("收到validate webhook创建请求")
	return UseValidate(w.logger, obj, w.client, ctx)
}

func (w *DeploymentWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
	w.logger.Info("收到validate webhook更新请求")
	return UseValidate(w.logger, newObj, w.client, ctx)
}

func (w *DeploymentWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	w.logger.Info("收到validate webhook删除请求")
	return nil
}

func UseValidate(logger logr.Logger, obj runtime.Object, myClient client.Client, ctx context.Context) error {
	//获取所有的资源
	var deploymetObjs appsv1.DeploymentList
	var statefulsetObjs appsv1.StatefulSetList
	var daemonsetObjs appsv1.DaemonSetList
	var meta v12.ObjectMeta
	var spec corev1.PodTemplateSpec
	switch obj.(type) {
	case *appsv1.Deployment:
		meta = obj.(*appsv1.Deployment).ObjectMeta
		spec = obj.(*appsv1.Deployment).Spec.Template
	case *appsv1.StatefulSet:
		meta = obj.(*appsv1.StatefulSet).ObjectMeta
		spec = obj.(*appsv1.StatefulSet).Spec.Template
	case *appsv1.DaemonSet:
		meta = obj.(*appsv1.DaemonSet).ObjectMeta
		spec = obj.(*appsv1.DaemonSet).Spec.Template
	}

	opts := client.ListOptions{
		Namespace: meta.Namespace,
	}
	err := myClient.List(ctx, &deploymetObjs, &opts)
	if err != nil {
		logger.Info("获取所有Deployment资源失败", "err", err)
		return err
	}
	err = myClient.List(ctx, &statefulsetObjs, &opts)
	if err != nil {
		logger.Info("获取所有StatefulSet资源失败", "err", err)
		return err
	}
	err = myClient.List(ctx, &daemonsetObjs, &opts)
	if err != nil {
		logger.Info("获取所有DaemonSet资源失败", "err", err)
		return err
	}

	var objsMap = make(map[string]runtime.Object)
	var objsReverseMap = make(map[string]*v12.ObjectMeta)
	//拼接资源map
	for _, v := range deploymetObjs.Items {
		objsMap[v.Name] = &v
		objsReverseMap[v.Name] = &v.ObjectMeta
	}
	for _, v := range statefulsetObjs.Items {
		objsMap[v.Name] = &v
		objsReverseMap[v.Name] = &v.ObjectMeta

	}
	for _, v := range daemonsetObjs.Items {
		objsMap[v.Name] = &v
		objsReverseMap[v.Name] = &v.ObjectMeta

	}

	//获取版本和依赖
	gVersion, deps, err := registry.GetVersionAndDependence(spec)
	if err != nil {
		logger.Info("获取版本和依赖失败", "err", err)
		return err
	}

	//检测依赖
	if err = registry.CheckForwardDependence(objsMap, deps); err != nil {
		logger.Info("检测正向依赖失败", "err", err)
		return err
	}
	if err = registry.CheckReverseDependence(objsReverseMap, meta.Name, gVersion); err != nil {
		logger.Info("检测反向依赖失败", "err", err)
		return err
	}
	return nil
}
func UseDefault(obj runtime.Object, logger logr.Logger) error {
	logger.Info("收到mutate webhook请求")
	var spec corev1.PodTemplateSpec
	var objN v12.ObjectMeta
	switch obj.(type) {
	case *appsv1.Deployment:
		spec = obj.(*appsv1.Deployment).Spec.Template
		objN = obj.(*appsv1.Deployment).ObjectMeta
	case *appsv1.StatefulSet:
		spec = obj.(*appsv1.StatefulSet).Spec.Template
		objN = obj.(*appsv1.StatefulSet).ObjectMeta
	case *appsv1.DaemonSet:
		spec = obj.(*appsv1.DaemonSet).Spec.Template
		objN = obj.(*appsv1.DaemonSet).ObjectMeta
	}
	gVersion, deps, err := registry.GetVersionAndDependence(spec)
	if err != nil {
		return err
	}
	//设置Annotation
	registry.SetObjVersion(&objN, gVersion, deps)
	for k, v := range deps {
		objN.Annotations[k+K8sAnnotationDependence] = v
	}
	return nil
}
