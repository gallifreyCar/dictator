package webhook

import (
	appsv1 "k8s.io/api/apps/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

//+kubebuilder:webhook:path=/mutate-apps-v1-daemonset,mutating=true,failurePolicy=fail,sideEffects=None,groups=apps,resources=daemonsets,verbs=create;update,versions=v1,name=mdaemonset.kb.io,admissionReviewVersions=v1
//+kubebuilder:webhook:path=/validate-apps-v1-daemonset,mutating=false,failurePolicy=fail,sideEffects=None,groups=apps,resources=daemonsets,verbs=create;update,versions=v1,name=vdaemonset.kb.io,admissionReviewVersions=v1

func SetupDaemonSetWebhookWithManager(mgr ctrl.Manager) error {
	hook := &Webhook{
		client: mgr.GetClient(),
		logger: logf.Log.WithName("[webhook.daemonset]"),
	}
	return ctrl.NewWebhookManagedBy(mgr).
		For(&appsv1.DaemonSet{}).
		WithDefaulter(hook).
		WithValidator(hook).
		Complete()
}
