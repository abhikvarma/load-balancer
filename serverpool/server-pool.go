package serverpool

import (
	"context"
	"errors"
	"time"

	"github.com/abhikvarma/load-balancer/backend"
	"github.com/abhikvarma/load-balancer/utils"

	"go.uber.org/zap"
)

type ServerPool interface {
	GetBackends() []backend.Backend
	GetNextValidPeer() backend.Backend
	AddBackend(backend.Backend)
	GetServerPoolSize() int
}

func HealthCheck(ctx context.Context, s ServerPool) {
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

func LaunchHealthCheck(ctx context.Context, sp ServerPool, health_check_interval_in_sec int) {
	t := time.NewTicker(time.Second * time.Duration(health_check_interval_in_sec))
	utils.Logger.Info("Starting health check...  ")
	for {
		select {
		case <-ctx.Done():
			utils.Logger.Info("Closing health check")
			return
		case <-t.C:
			go HealthCheck(ctx, sp)
		}
	}
}

func NewServerPool(strategy utils.LBStrategy) (ServerPool, error) {
	switch strategy {
	case utils.RoundRobin:
		return &roundRobinServerPool{
			backends: make([]backend.Backend, 0),
			current:  0,
		}, nil
	case utils.LeastConnected:
		return &lcServerPool{
			backends: make([]backend.Backend, 0),
		}, nil
	default:
		return nil, errors.New("invalid strategy")
	}
}
