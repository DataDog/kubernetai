package kubernetai

import (
	"reflect"

	"github.com/coredns/coredns/plugin/kubernetes"
	"github.com/coredns/coredns/plugin/kubernetes/object"
)

type podHandlerItf interface {
	PodWithIP(k *kubernetes.Kubernetes, ip string) *object.Pod
	IsPodVerified(k *kubernetes.Kubernetes) bool
}

type podHandler struct {
	IsVerifiedCache map[string]podVerified
}

type podVerified struct {
	isVerified  bool
	initialized bool
}

// podWithIP return the api.Pod for source IP ip. It returns nil if nothing can be found.
func (p *podHandler) PodWithIP(k *kubernetes.Kubernetes, ip string) *object.Pod {
	ps := k.APIConn.PodIndex(ip)
	if len(ps) == 0 {
		return nil
	}
	return ps[0]
}

func (p *podHandler) IsPodVerified(k *kubernetes.Kubernetes) bool {
	cacheKey := k.Name()
	cacheVal := p.IsVerifiedCache[cacheKey]
	if cacheVal.initialized == false {
		kubeReflection := reflect.ValueOf(k).Elem()
		podMode := kubeReflection.FieldByName("podMode")
		cacheVal.isVerified = !(podMode.String() == "disabled")
		cacheVal.initialized = true
	}
	return cacheVal.isVerified
}
