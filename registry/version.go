package registry

import (
	"errors"
	"fmt"
	"github.com/Masterminds/semver/v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	_ "net/http"
	"strings"
)

// 获取版本
// 从init容器和普通容器中依次遍历, 找到第一个符合语义化版本的镜像tag
func getVersionByPodTemplate(podSpec *corev1.PodTemplateSpec) string {
	containers := make([]corev1.Container, 0, len(podSpec.Spec.InitContainers)+len(podSpec.Spec.Containers))
	containers = append(containers, podSpec.Spec.InitContainers...)
	containers = append(containers, podSpec.Spec.Containers...)
	for _, c := range containers {
		i := strings.LastIndexByte(c.Image, ':')
		if i == -1 {
			continue
		}
		v, err := semver.NewVersion(c.Image[i+1:])
		if err != nil {
			continue
		}
		return fmt.Sprintf("%d.%d.%d", v.Major(), v.Minor(), v.Patch())
	}

	return ""
}

// 获取依赖约束
// 从init容器和普通容器中依次遍历, 获取每个镜像的依赖约束
func getDependenceByPodTemplate(podSpec *corev1.PodTemplateSpec) (map[string]string, error) {
	deps := make(map[string]string)

	containers := make([]corev1.Container, 0, len(podSpec.Spec.InitContainers)+len(podSpec.Spec.Containers))
	containers = append(containers, podSpec.Spec.InitContainers...)
	containers = append(containers, podSpec.Spec.Containers...)
	for _, c := range containers {
		i := strings.LastIndexByte(c.Image, ':')
		if i == -1 {
			continue
		}

		dependence, err := GetImageDependenceRaw(c.Image)
		if err != nil {
			return nil, err
		}
		for k, v := range dependence {
			if got, ok := deps[k]; ok {
				deps[k] = got + "," + v
			} else {
				deps[k] = v
			}
		}
	}

	return deps, nil
}

// GetVersionAndDependence 从远程私人仓库获取版本和依赖约束
func GetVersionAndDependence(podSpec corev1.PodTemplateSpec) (string, map[string]string, error) {
	version := getVersionByPodTemplate(&podSpec)
	deps, err := getDependenceByPodTemplate(&podSpec)
	return version, deps, err
}

func CheckForwardDependence(objs map[string]runtime.Object, deps map[string]string) error {
	klog.V(4).Infof("正向依赖检查: %v\n", deps)
	for svc, constraint := range deps {
		c, err := semver.NewConstraint(constraint)
		if err != nil {
			return err
		}

		obj := objs[svc]
		if obj == nil {
			klog.V(4).Info("被依赖的服务不存在: %s\n", svc)
			continue
		}

		version, err := GetVersion(obj)
		if version == "" {
			klog.V(4).Infof("被依赖的服务版本为空: %s\n", svc)
			continue
		}

		v, err := semver.NewVersion(version)
		if err != nil {
			return err
		}
		if !c.Check(v) {
			return errors.New(fmt.Sprintf("正向依赖检查失败，%s版本(%s)不符合依赖约束(%s)", svc, version, constraint))
		}
	}
	return nil
}

func CheckReverseDependence(objs map[string]*v12.ObjectMeta, svc string, version string) error {
	klog.V(4).Infof("反向依赖检查: %s %s\n", svc, version)
	if version == "" {
		return nil
	}

	v, err := semver.NewVersion(version)
	if err != nil {
		return err
	}

	key := svc + K8sAnnotationDependence
	for _, obj := range objs {
		depRaw := obj.GetAnnotations()[key]
		if depRaw == "" {
			continue
		}
		deps := strings.Split(depRaw, ",")
		for _, dep := range deps {
			c, err := semver.NewConstraint(dep)
			if err != nil {
				return err
			}
			if !c.Check(v) {
				return errors.New(fmt.Sprintf("反向依赖检查失败，%s版本(%s)不符合%s的依赖约束(%s)", svc, version, obj.GetName(), dep))
			}
		}
	}
	return nil
}

// SetObjVersion 设置对象的版本号
func SetObjVersion(obj *v12.ObjectMeta, version string, deps map[string]string) {
	Labels := obj.GetLabels()
	if Labels == nil {
		Labels = map[string]string{}
	}
	Labels[K8sLabelVersion] = version
	obj.SetLabels(Labels)

	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	for k, v := range deps {
		annotations[k+K8sAnnotationDependence] = v
	}
	obj.SetAnnotations(annotations)
}

func GetVersion(obj runtime.Object) (string, error) {
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

	version := objN.GetLabels()[K8sLabelVersion]
	if version != "" {
		return version, nil
	}

	return getVersionByPodTemplate(&spec), nil

}
