package loadbalancer

import (
	"context"
	"net/http"
	"time"

	"github.com/abhikvarma/load-balancer/backend"
	"github.com/abhikvarma/load-balancer/serverpool"
	"github.com/abhikvarma/load-balancer/utils"
	"go.uber.org/zap"
)

type key int

type LoadBalancer interface {
	Serve(http.ResponseWriter, *http.Request)
	LaunchHealthCheck(context.Context, int)
}

type loadBalancer struct {
	serverpool serverpool.ServerPool
}

const RETRIES_ATTEMPTED key = 0

func AllowRetry(r *http.Request, max_retries int) (int, bool) {
	attempts, ok := r.Context().Value(RETRIES_ATTEMPTED).(int)
	if ok && attempts > max_retries {
		return attempts, false
	}
	return attempts, true
}

func (lb *loadBalancer) Serve(w http.ResponseWriter, r *http.Request) {
	peer := lb.serverpool.GetNextValidPeer()
	if peer != nil {
		peer.Serve(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

func (lb *loadBalancer) LaunchHealthCheck(ctx context.Context, health_check_interval_in_sec int) {
	t := time.NewTicker(time.Second * time.Duration(health_check_interval_in_sec))
	utils.Logger.Info("Starting health check...  ")
	for {
		select {
		case <-ctx.Done():
			utils.Logger.Info("Closing health check")
			return
		case <-t.C:
			go healthCheck(ctx, lb.serverpool)
		}
	}
}

func healthCheck(ctx context.Context, s serverpool.ServerPool) {
	aliveChannel := make(chan bool, 1)

	for _, b := range s.GetBackends() {
		b := b
		requestCtx, stop := context.WithTimeout(ctx, 10*time.Second)
		defer stop()
		status := "up"

		go backend.IsBackendAlive(requestCtx, aliveChannel, b.GetURL())

		select {
		case <-ctx.Done():
			utils.Logger.Info("Gracefully shutting down health check")
			return
		case alive := <-aliveChannel:
			b.SetAlive(alive)
			if !alive {
				status = "down"
			}
		}
		utils.Logger.Debug(
			"URL Status",
			zap.String("URL", b.GetURL().String()),
			zap.String("status", status),
		)
	}
}

func NewLoadBalancer(sp serverpool.ServerPool) LoadBalancer {
	return &loadBalancer{
		serverpool: sp,
	}
}
