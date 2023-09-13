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

//+kubebuilder:webhook:path=/mutate-apps-v1-daemonset,mutating=true,failurePolicy=fail,sideEffects=None,groups=apps,resources=daemonsets,verbs=create;update,versions=v1,name=mdaemonset.kb.io,admissionReviewVersions=v1
//+kubebuilder:webhook:path=/validate-apps-v1-daemonset,mutating=false,failurePolicy=fail,sideEffects=None,groups=apps,resources=daemonsets,verbs=create;update,versions=v1,name=vdaemonset.kb.io,admissionReviewVersions=v1

type DaemonSetWebhook struct {
	client client.Client
	logger logr.Logger
}

func SetupDaemonSetWebhookWithManager(mgr ctrl.Manager) error {
	hook := &DaemonSetWebhook{
		client: mgr.GetClient(),
		logger: logf.Log.WithName("[webhook.daemonset]"),
	}
	return ctrl.NewWebhookManagedBy(mgr).
		For(&appsv1.DaemonSet{}).
		WithDefaulter(hook).
		WithValidator(hook).
		Complete()
}

func (d DaemonSetWebhook) Default(ctx context.Context, obj runtime.Object) error {
	return UseDefault(obj, d.logger)
}

func (d DaemonSetWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	d.logger.Info("收到validate webhook创建请求")
	return UseValidate(d.logger, obj, d.client, ctx)
}

func (d DaemonSetWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
	d.logger.Info("收到validate webhook更新请求")
	return UseValidate(d.logger, newObj, d.client, ctx)
}

func (d DaemonSetWebhook) ValidateDelete(_ context.Context, _ runtime.Object) error { return nil }
