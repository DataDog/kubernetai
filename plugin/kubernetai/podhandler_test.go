package kubernetai

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/coredns/coredns/plugin/kubernetes"
)

func Test_podHandler_IsPodVerified(t *testing.T) {
	type fields struct {
		IsVerifiedCache map[string]podVerified
	}
	type args struct {
		k *kubernetes.Kubernetes
	}

	// Reflexion magic to specify podMode.
	k := kubernetes.New([]string{"cluster.local"})
	kubeReflection := reflect.ValueOf(k).Elem()
	podMode := kubeReflection.FieldByName("podMode")
	podMode = reflect.NewAt(podMode.Type(), unsafe.Pointer(podMode.UnsafeAddr())).Elem()
	podMode.SetString("verified")

	cacheMap := make(map[string]podVerified, 3)
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "Should return false on podModeDisabled",
			fields: fields{IsVerifiedCache: cacheMap},
			args: args{
				k: kubernetes.New([]string{"cluster.local"}),
			},
			want: false,
		},
		{
			name:   "Should return true on podModeVerified",
			fields: fields{IsVerifiedCache: cacheMap},
			args: args{
				k: k,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &podHandler{
				IsVerifiedCache: tt.fields.IsVerifiedCache,
			}
			if got := p.IsPodVerified(tt.args.k); got != tt.want {
				t.Errorf("podHandler.IsPodVerified() = %v, want %v", got, tt.want)
			}
		})
	}
}
