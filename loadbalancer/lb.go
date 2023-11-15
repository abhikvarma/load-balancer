package loadbalancer

import (
	"net/http"

	"github.com/abhikvarma/load-balancer/serverpool"
)

type key int

const RETRIES_ATTEMPTED key = 0

func AllowRetry(r *http.Request, max_retries int) (int, bool) {
	attempts, ok := r.Context().Value(RETRIES_ATTEMPTED).(int)
	if ok && attempts > max_retries {
		return attempts, false
	}
	return attempts, true
}

type LoadBalancer interface {
	Serve(http.ResponseWriter, *http.Request)
}

type loadBalancer struct {
	serverpool serverpool.ServerPool
}

func (lb *loadBalancer) Serve(w http.ResponseWriter, r *http.Request) {
	peer := lb.serverpool.GetNextValidPeer()
	if peer != nil {
		peer.Serve(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

func NewLoadBalancer(sp serverpool.ServerPool) LoadBalancer {
	return &loadBalancer{
		serverpool: sp,
	}
}
