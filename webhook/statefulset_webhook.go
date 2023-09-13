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

//+kubebuilder:webhook:path=/mutate-apps-v1-statefulset,mutating=true,failurePolicy=fail,sideEffects=None,groups=apps,resources=statefulsets,verbs=create;update,versions=v1,name=mstatefulset.kb.io,admissionReviewVersions=v1
//+kubebuilder:webhook:path=/validate-apps-v1-statefulset,mutating=false,failurePolicy=fail,sideEffects=None,groups=apps,resources=statefulsets,verbs=create;update,versions=v1,name=vstatefulset.kb.io,admissionReviewVersions=v1

type StatefulSetWebhook struct {
	client client.Client
	logger logr.Logger
}

func SetupStatefulSetWebhookWithManager(mgr ctrl.Manager) error {
	hook := &StatefulSetWebhook{
		client: mgr.GetClient(),
		logger: logf.Log.WithName("[webhook.statefulset]"),
	}
	return ctrl.NewWebhookManagedBy(mgr).
		For(&appsv1.StatefulSet{}).
		WithDefaulter(hook).
		WithValidator(hook).
		Complete()
}

func (s StatefulSetWebhook) Default(ctx context.Context, obj runtime.Object) error {
	return UseDefault(obj, s.logger)
}

func (s StatefulSetWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	s.logger.Info("收到validate webhook创建请求")
	return UseValidate(s.logger, obj, s.client, ctx)
}

func (s StatefulSetWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
	s.logger.Info("收到validate webhook更新请求")
	return UseValidate(s.logger, newObj, s.client, ctx)
}

func (s StatefulSetWebhook) ValidateDelete(_ context.Context, _ runtime.Object) error { return nil }
