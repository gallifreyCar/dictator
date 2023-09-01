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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

func (w *DeploymentWebhook) Default(ctx context.Context, obj runtime.Object) error {
	w.logger.Info("收到mutate webhook请求")
	gVersion, deps, err := registry.GetVersionAndDependence(registry.KRTDeployment, obj.(*unstructured.Unstructured))
	if err != nil {
		return err
	}
	//设置Annotation
	registry.SetObjVersion(obj.(*unstructured.Unstructured), gVersion, deps)
	obj.(*appsv1.Deployment).Annotations["dictator.wellcloud.cc/annotation"] = "dictator"

	return nil
}

func (w *DeploymentWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	w.logger.Info("收到validate webhook创建请求")
	//获取所有的资源
	objs := unstructured.Unstructured{}
	opts := client.ListOptions{
		Namespace: "wellcloud",
	}
	err := w.client.List(ctx, &objs, &opts)

	var objsMap = make(map[string]*unstructured.Unstructured)
	//拼接资源map
	for k, v := range objs.Object {
		objsMap[k] = v.(*unstructured.Unstructured)
	}

	//获取版本和依赖
	gVersion, deps, err := registry.GetVersionAndDependence(registry.KRTDeployment, obj.(*unstructured.Unstructured))
	if err != nil {
		return err
	}

	//检测依赖
	if err = registry.CheckForwardDependence(objsMap, deps); err != nil {
		return err
	}
	if err = registry.CheckReverseDependence(objsMap, obj.GetObjectKind().GroupVersionKind().Kind, gVersion); err != nil {
		return err
	}
	return nil
}

func (w *DeploymentWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
	w.logger.Info("收到validate webhook更新请求")
	//获取所有的资源
	objs := unstructured.Unstructured{}
	opts := client.ListOptions{
		Namespace: "wellcloud",
	}
	err := w.client.List(ctx, &objs, &opts)

	var objsMap = make(map[string]*unstructured.Unstructured)
	//拼接资源map
	for k, v := range objs.Object {
		objsMap[k] = v.(*unstructured.Unstructured)
	}

	//获取版本和依赖
	gVersion, deps, err := registry.GetVersionAndDependence(registry.KRTDeployment, oldObj.(*unstructured.Unstructured))
	if err != nil {
		return err
	}

	//检测依赖
	if err = registry.CheckForwardDependence(objsMap, deps); err != nil {
		return err
	}
	if err = registry.CheckReverseDependence(objsMap, oldObj.GetObjectKind().GroupVersionKind().Kind, gVersion); err != nil {
		return err
	}
	return nil
	return nil
}

func (w *DeploymentWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	w.logger.Info("收到validate webhook删除请求")
	return nil
}
