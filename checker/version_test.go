package checker

import "testing"

func Test_completeImageRegistry(t *testing.T) {
	type args struct {
		address string
		part    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "test", args: args{
			address: "harbor:5000",
			part:    "wecloud/wmc:1.5.1",
		}, want: "harbor:5000/wecloud/wmc:1.5.1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := completeImageRegistry(tt.args.address, tt.args.part); got != tt.want {
				t.Errorf("completeImageRegistry() = %v, want %v", got, tt.want)
			}
		})
	}
}
