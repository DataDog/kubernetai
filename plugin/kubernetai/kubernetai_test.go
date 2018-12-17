package kubernetai

import (
	"github.com/coredns/coredns/plugin/kubernetes"
	"github.com/coredns/coredns/plugin/kubernetes/object"
	"github.com/miekg/dns"
	"net"
	"reflect"
	"testing"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
)

var podip string
var podCache bool

type k8iPodHandlerTester struct{}

var k8iPodHandlerTest k8iPodHandlerTester

func (k8i *k8iPodHandlerTester) PodWithIP(k *kubernetes.Kubernetes, ip string) *object.Pod {
	if ip == "" {
		return nil
	}
	pod := &object.Pod{
		Namespace: "test-1",
		PodIP:     ip,
	}
	return pod
}

func (k8i *k8iPodHandlerTester) IsPodVerified(k *kubernetes.Kubernetes) bool {
	return podCache
}

type responseWriterTest struct {
	dns.ResponseWriter
}

func (res *responseWriterTest) RemoteAddr() net.Addr {
	ip := net.ParseIP(podip)
	return &net.UDPAddr{
		IP:   ip,
		Port: 53,
	}
}

func TestKubernetai_AutoPath(t *testing.T) {
	type fields struct {
		Zones          []string
		Next           plugin.Handler
		Kubernetes     []*kubernetes.Kubernetes
		autoPathSearch []string
		p              *k8iPodHandlerTester
	}
	type args struct {
		state request.Request
	}

	w := &responseWriterTest{}

	k8sClusterLocal := &kubernetes.Kubernetes{
		Zones: []string{
			"cluster.local.",
		},
	}
	k8sFlusterLocal := &kubernetes.Kubernetes{
		Zones: []string{
			"fluster.local.",
		},
	}
	defaultK8iConfig := fields{
		Kubernetes: []*kubernetes.Kubernetes{
			k8sFlusterLocal,
			k8sClusterLocal,
		},
		p: &k8iPodHandlerTest,
	}

	tests := []struct {
		name     string
		fields   fields
		args     args
		want     []string
		ip       string
		podCache bool
	}{
		{
			name:   "standard autopath cluster.local",
			fields: defaultK8iConfig,
			args: args{
				state: request.Request{
					W: w,
					Req: &dns.Msg{
						Question: []dns.Question{
							{Name: "svc-1-a.test-1.svc.cluster.local.", Qtype: 1, Qclass: 1},
						},
					},
				},
			},
			want:     []string{"test-1.svc.cluster.local.", "svc.cluster.local.", "cluster.local.", "test-1.svc.fluster.local.", "svc.fluster.local.", "fluster.local.", ""},
			ip:       "172.17.0.7",
			podCache: true,
		},
		{
			name:   "standard autopath with podMode = podModeDisabled on cluster.local",
			fields: defaultK8iConfig,
			args: args{
				state: request.Request{
					W: w,
					Req: &dns.Msg{
						Question: []dns.Question{
							{Name: "svc-1-a.test-1.svc.cluster.local.", Qtype: 1, Qclass: 1},
						},
					},
				},
			},
			want:     []string{"svc.cluster.local.", "cluster.local.", "svc.fluster.local.", "fluster.local.", ""},
			ip:       "172.17.0.7",
			podCache: false,
		},
		{
			name:   "standard autopath servicename.svc",
			fields: defaultK8iConfig,
			args: args{
				state: request.Request{
					W: w,
					Req: &dns.Msg{
						Question: []dns.Question{
							{Name: "svc-2-a.test-2.test-1.svc.cluster.local.", Qtype: 1, Qclass: 1},
						},
					},
				},
			},
			want:     []string{"test-1.svc.cluster.local.", "svc.cluster.local.", "cluster.local.", "test-1.svc.fluster.local.", "svc.fluster.local.", "fluster.local.", ""},
			ip:       "172.17.0.7",
			podCache: true,
		},
		{
			name:   "standard autopath servicename.svc with podMode = podModeDisabled",
			fields: defaultK8iConfig,
			args: args{
				state: request.Request{
					W: w,
					Req: &dns.Msg{
						Question: []dns.Question{
							{Name: "svc-2-a.test-2.test-1.svc.cluster.local.", Qtype: 1, Qclass: 1},
						},
					},
				},
			},
			want:     []string{"svc.cluster.local.", "cluster.local.", "svc.fluster.local.", "fluster.local.", ""},
			ip:       "172.17.0.7",
			podCache: false,
		},
		{
			name:   "standard autopath lookup fluster in cluster.local",
			fields: defaultK8iConfig,
			args: args{
				state: request.Request{
					W: w,
					Req: &dns.Msg{
						Question: []dns.Question{
							{Name: "svc-d.test-2.svc.fluster.local.svc.cluster.local.", Qtype: 1, Qclass: 1},
						},
					},
				},
			},
			want:     []string{"test-1.svc.cluster.local.", "svc.cluster.local.", "cluster.local.", "test-1.svc.fluster.local.", "svc.fluster.local.", "fluster.local.", ""},
			ip:       "172.17.0.7",
			podCache: true,
		},
		{
			name:   "standard autopath lookup fluster in cluster.local with podMode = podModeDisabled",
			fields: defaultK8iConfig,
			args: args{
				state: request.Request{
					W: w,
					Req: &dns.Msg{
						Question: []dns.Question{
							{Name: "svc-d.test-2.svc.fluster.local.svc.cluster.local.", Qtype: 1, Qclass: 1},
						},
					},
				},
			},
			want:     []string{"svc.cluster.local.", "cluster.local.", "svc.fluster.local.", "fluster.local.", ""},
			ip:       "172.17.0.7",
			podCache: false,
		},
		{
			name:   "not in zone",
			fields: defaultK8iConfig,
			args: args{
				state: request.Request{
					W: w,
					Req: &dns.Msg{
						Question: []dns.Question{
							{Name: "svc-1-a.test-1.svc.zone.local.", Qtype: 1, Qclass: 1},
						},
					},
				},
			},
			ip:       "172.17.0.7",
			want:     nil,
			podCache: true,
		},
		{
			name:   "requesting pod does not exist",
			fields: defaultK8iConfig,
			args: args{
				state: request.Request{
					W: w,
					Req: &dns.Msg{
						Question: []dns.Question{
							{Name: "svc-1-a.test-1.svc.zone.local.", Qtype: 1, Qclass: 1},
						},
					},
				},
			},
			ip:       "",
			want:     nil,
			podCache: true,
		},
		{
			name:   "AXFR should return nil",
			fields: defaultK8iConfig,
			args: args{
				state: request.Request{
					W: w,
					Req: &dns.Msg{
						Question: []dns.Question{
							{Name: "zone.local.", Qtype: dns.TypeAXFR, Qclass: 1},
						},
					},
				},
			},
			ip:       "",
			want:     nil,
			podCache: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k8i := Kubernetai{
				Zones:          tt.fields.Zones,
				Next:           tt.fields.Next,
				Kubernetes:     tt.fields.Kubernetes,
				autoPathSearch: tt.fields.autoPathSearch,
				p:              tt.fields.p,
			}
			podCache = tt.podCache
			podip = tt.ip
			if got := k8i.AutoPath(tt.args.state); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Kubernetai.AutoPath() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
