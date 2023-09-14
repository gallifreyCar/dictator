package webhook

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"gitlab.wellcloud.cc/cloud/dictator/checker"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
)

var myChecker checker.MyChecker

type Webhook struct {
	client client.Client
	logger logr.Logger
}

func UseDefault(w *Webhook, obj runtime.Object, checker checker.Chercker) error {
	w.logger.Info("收到mutate webhook请求")
	var spec corev1.PodTemplateSpec
	var meta metav1.ObjectMeta
	switch obj.(type) {
	case *appsv1.Deployment:
		spec = obj.(*appsv1.Deployment).Spec.Template
		meta = obj.(*appsv1.Deployment).ObjectMeta
	case *appsv1.StatefulSet:
		spec = obj.(*appsv1.StatefulSet).Spec.Template
		meta = obj.(*appsv1.StatefulSet).ObjectMeta
	case *appsv1.DaemonSet:
		spec = obj.(*appsv1.DaemonSet).Spec.Template
		meta = obj.(*appsv1.DaemonSet).ObjectMeta
	default:
		return fmt.Errorf("不支持的资源类型: %s", obj.GetObjectKind().GroupVersionKind())
	}
	gVersion, deps, err := checker.GetVersionAndDependence(spec)
	if err != nil {
		return err
	}
	//设置Annotation
	checker.SetObjVersion(&meta, gVersion, deps)
	return nil
}

func UseValidate(w *Webhook, obj runtime.Object, ctx context.Context, ck checker.Chercker) error {
	//获取所有的资源
	var deploymetObjs appsv1.DeploymentList
	var statefulsetObjs appsv1.StatefulSetList
	var daemonsetObjs appsv1.DaemonSetList
	var meta *metav1.ObjectMeta
	var anno map[string]string
	deps := make(map[string]string)

	switch obj.(type) {
	case *appsv1.Deployment:
		meta = &obj.(*appsv1.Deployment).ObjectMeta
		anno = obj.(*appsv1.Deployment).GetAnnotations()
	case *appsv1.StatefulSet:
		meta = &obj.(*appsv1.StatefulSet).ObjectMeta
		anno = obj.(*appsv1.StatefulSet).GetAnnotations()
	case *appsv1.DaemonSet:
		meta = &obj.(*appsv1.DaemonSet).ObjectMeta
		anno = obj.(*appsv1.DaemonSet).GetAnnotations()
	default:
		return fmt.Errorf("不支持的资源类型: %s", obj.GetObjectKind().GroupVersionKind())
	}

	opts := client.ListOptions{
		Namespace: meta.Namespace,
	}
	err := w.client.List(ctx, &deploymetObjs, &opts)
	if err != nil {
		w.logger.Info("获取所有Deployment资源失败", "err", err)
		return err
	}
	err = w.client.List(ctx, &statefulsetObjs, &opts)
	if err != nil {
		w.logger.Info("获取所有StatefulSet资源失败", "err", err)
		return err
	}
	err = w.client.List(ctx, &daemonsetObjs, &opts)
	if err != nil {
		w.logger.Info("获取所有DaemonSet资源失败", "err", err)
		return err
	}

	var objsMap = make(map[string]runtime.Object)
	var objsReverseMap = make(map[string]*metav1.ObjectMeta)
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
	gVersion, err := ck.GetVersion(obj)
	if err != nil {
		w.logger.Info("获取版本失败", "err", err)
		return err
	}
	for k, v := range anno {
		if strings.HasSuffix(k, checker.K8sAnnotationDependence) {
			dk := strings.TrimSuffix(k, checker.K8sAnnotationDependence)
			deps[dk] = v
		}
	}

	//检测依赖
	if err = ck.CheckForwardDependence(objsMap, deps, w.logger); err != nil {
		w.logger.Info("检测正向依赖失败", "err", err)
		return err
	}
	if err = ck.CheckReverseDependence(objsReverseMap, meta.Name, gVersion, w.logger); err != nil {
		w.logger.Info("检测反向依赖失败", "err", err)
		return err
	}
	return nil
}

func (w *Webhook) Default(_ context.Context, obj runtime.Object) error {
	return UseDefault(w, obj, &myChecker)
}

func (w *Webhook) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	w.logger.Info("收到validate webhook创建请求")
	return UseValidate(w, obj, ctx, &myChecker)
}

func (w *Webhook) ValidateUpdate(ctx context.Context, _, newObj runtime.Object) error {
	w.logger.Info("收到validate webhook更新请求")
	return UseValidate(w, newObj, ctx, &myChecker)
}

func (w *Webhook) ValidateDelete(_ context.Context, _ runtime.Object) error { return nil }

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps,resources=deployments/finalizers,verbs=update

//+kubebuilder:webhook:path=/mutate-apps-v1-deployment,mutating=true,failurePolicy=fail,sideEffects=None,groups=apps,resources=deployments,verbs=create;update,versions=v1,name=mdeployment.kb.io,admissionReviewVersions=v1
//+kubebuilder:webhook:path=/validate-apps-v1-deployment,mutating=false,failurePolicy=fail,sideEffects=None,groups=apps,resources=deployments,verbs=create;update,versions=v1,name=vdeployment.kb.io,admissionReviewVersions=v1

func SetupDeploymentWebhookWithManager(mgr ctrl.Manager) error {
	hook := &Webhook{
		client: mgr.GetClient(),
		logger: logf.Log.WithName("[webhook.deployment]"),
	}
	return ctrl.NewWebhookManagedBy(mgr).
		For(&appsv1.Deployment{}).
		WithDefaulter(hook).
		WithValidator(hook).
		Complete()
}

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

//+kubebuilder:webhook:path=/mutate-apps-v1-statefulset,mutating=true,failurePolicy=fail,sideEffects=None,groups=apps,resources=statefulsets,verbs=create;update,versions=v1,name=mstatefulset.kb.io,admissionReviewVersions=v1
//+kubebuilder:webhook:path=/validate-apps-v1-statefulset,mutating=false,failurePolicy=fail,sideEffects=None,groups=apps,resources=statefulsets,verbs=create;update,versions=v1,name=vstatefulset.kb.io,admissionReviewVersions=v1

func SetupStatefulSetWebhookWithManager(mgr ctrl.Manager) error {
	hook := &Webhook{
		client: mgr.GetClient(),
		logger: logf.Log.WithName("[webhook.statefulset]"),
	}
	return ctrl.NewWebhookManagedBy(mgr).
		For(&appsv1.StatefulSet{}).
		WithDefaulter(hook).
		WithValidator(hook).
		Complete()
}
