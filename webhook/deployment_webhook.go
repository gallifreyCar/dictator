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
	appsv1 "k8s.io/api/apps/v1"
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

func (w *DeploymentWebhook) ValidateDelete(_ context.Context, _ runtime.Object) error { return nil }
