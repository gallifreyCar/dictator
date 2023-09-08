package webhook

import (
	"context"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func TestDeploymentWebhook_Default(t *testing.T) {
	type fields struct {
		client client.Client
		logger logr.Logger
	}
	type args struct {
		ctx context.Context
		obj runtime.Object
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "test", fields: fields{}, args: args{
			ctx: context.Background(),
			obj: &v1.Deployment{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       v1.DeploymentSpec{},
				Status:     v1.DeploymentStatus{},
			},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &DeploymentWebhook{
				client: tt.fields.client,
				logger: tt.fields.logger,
			}
			if err := w.Default(tt.args.ctx, tt.args.obj); (err != nil) != tt.wantErr {
				t.Errorf("Default() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeploymentWebhook_ValidateCreate(t *testing.T) {
	type fields struct {
		client client.Client
		logger logr.Logger
	}
	type args struct {
		ctx context.Context
		obj runtime.Object
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test", fields: fields{}, args: args{
				ctx: context.Background(),
				obj: &v1.Deployment{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec:       v1.DeploymentSpec{},
					Status:     v1.DeploymentStatus{},
				},
			}, wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &DeploymentWebhook{
				client: tt.fields.client,
				logger: tt.fields.logger,
			}
			if err := w.ValidateCreate(tt.args.ctx, tt.args.obj); (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeploymentWebhook_ValidateUpdate(t *testing.T) {
	type fields struct {
		client client.Client
		logger logr.Logger
	}
	type args struct {
		ctx    context.Context
		oldObj runtime.Object
		newObj runtime.Object
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "test", fields: fields{}, args: args{
			ctx: context.Background(),
			oldObj: &v1.Deployment{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       v1.DeploymentSpec{},
				Status:     v1.DeploymentStatus{},
			},
			newObj: &v1.Deployment{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       v1.DeploymentSpec{},
				Status:     v1.DeploymentStatus{},
			},
		}, wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &DeploymentWebhook{
				client: tt.fields.client,
				logger: tt.fields.logger,
			}
			if err := w.ValidateUpdate(tt.args.ctx, tt.args.oldObj, tt.args.newObj); (err != nil) != tt.wantErr {
				t.Errorf("ValidateUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
