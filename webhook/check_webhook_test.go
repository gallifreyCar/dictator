package webhook

import (
	"context"
	"github.com/go-logr/logr"
	"gitlab.wellcloud.cc/cloud/dictator/checker"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

func TestUseDefault(t *testing.T) {
	type args struct {
		w       *Webhook
		obj     runtime.Object
		checker checker.Chercker
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "webhook",
			args: args{
				w: &Webhook{
					client: &myFakeClient{},
					logger: logf.Log.WithName("[webhook]"),
				},
				obj:     &appsv1.Deployment{},
				checker: &myFakeChecker{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := UseDefault(tt.args.w, tt.args.obj, tt.args.checker); (err != nil) != tt.wantErr {
				t.Errorf("UseDefault() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUseValidate(t *testing.T) {
	type args struct {
		w   *Webhook
		obj runtime.Object
		ctx context.Context
		ck  checker.Chercker
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "webhook",
			args: args{
				w: &Webhook{
					client: &myFakeClient{},
					logger: logf.Log.WithName("[webhook]"),
				},
				obj: &appsv1.Deployment{

					ObjectMeta: metav1.ObjectMeta{
						Namespace:   "default",
						Annotations: map[string]string{},
					},
				},
				ctx: context.Background(),
				ck:  &myFakeChecker{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := UseValidate(tt.args.w, tt.args.obj, tt.args.ctx, tt.args.ck); (err != nil) != tt.wantErr {
				t.Errorf("UseValidate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

var _ client.Client = (*myFakeClient)(nil)

type myFakeClient struct {
}

func (m *myFakeClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	//TODO implement me
	panic("implement me")
}

func (m *myFakeClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	list = &appsv1.DeploymentList{
		Items: []appsv1.Deployment{
			{},
		},
	}
	return nil
}

func (m *myFakeClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	//TODO implement me
	panic("implement me")
}

func (m *myFakeClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	//TODO implement me
	panic("implement me")
}

func (m myFakeClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	//TODO implement me
	panic("implement me")
}

func (m myFakeClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	//TODO implement me
	panic("implement me")
}

func (m myFakeClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	//TODO implement me
	panic("implement me")
}

func (m myFakeClient) Status() client.StatusWriter {
	//TODO implement me
	panic("implement me")
}

func (m myFakeClient) Scheme() *runtime.Scheme {
	//TODO implement me
	panic("implement me")
}

func (m myFakeClient) RESTMapper() meta.RESTMapper {
	//TODO implement me
	panic("implement me")
}

var _ checker.Chercker = (*myFakeChecker)(nil)

type myFakeChecker struct{}

func (m *myFakeChecker) GetVersion(obj runtime.Object) (string, error) {
	return "v1", nil
}

func (m *myFakeChecker) GetVersionAndDependence(podSpec corev1.PodTemplateSpec) (string, map[string]string, error) {
	return "v1", map[string]string{
		"app": "v1",
		"db":  "v1",
	}, nil
}

func (m *myFakeChecker) CheckForwardDependence(objs map[string]runtime.Object, deps map[string]string, logger logr.Logger) error {
	return nil
}

func (m *myFakeChecker) CheckReverseDependence(objs map[string]*v12.ObjectMeta, svc string, version string, logger logr.Logger) error {
	return nil
}

func (m *myFakeChecker) SetObjVersion(meta *v12.ObjectMeta, version string, deps map[string]string) {
	return
}
