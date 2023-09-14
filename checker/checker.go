package checker

import (
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Chercker interface {
	GetVersion(obj runtime.Object) (string, error)
	GetVersionAndDependence(podSpec corev1.PodTemplateSpec) (string, map[string]string, error)
	CheckForwardDependence(objs map[string]runtime.Object, deps map[string]string, logger logr.Logger) error
	CheckReverseDependence(objs map[string]*v12.ObjectMeta, svc string, version string, logger logr.Logger) error
	SetObjVersion(meta *v12.ObjectMeta, version string, deps map[string]string)
}
