package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/abhikvarma/load-balancer/backend"
	"github.com/abhikvarma/load-balancer/loadbalancer"
	"github.com/abhikvarma/load-balancer/serverpool"
	"github.com/abhikvarma/load-balancer/utils"

	"go.uber.org/zap"
)

func main() {
	logger := utils.InitLogger()
	defer logger.Sync()

	config, err := utils.GetLBConfig()
	if err != nil {
		utils.Logger.Fatal(err.Error())
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sp, err := serverpool.NewServerPool(utils.GetLBStrategy(config.Strategy))
	if err != nil {
		utils.Logger.Fatal(err.Error())
	}
	lb := loadbalancer.NewLoadBalancer(sp)

	for _, b := range config.Backends {
		endpoint, err := url.Parse(b)
		if err != nil {
			utils.Logger.Fatal(err.Error())
		}

		rp := httputil.NewSingleHostReverseProxy(endpoint)
		backendServer := backend.NewBackend(endpoint, rp)
		rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
			logger.Info(fmt.Sprintf("error handling the request %s", r.RequestURI),
				zap.String("host", endpoint.Host),
				zap.Error(e),
			)

			attempts, allow := loadbalancer.AllowRetry(r, config.MaxRetries)
			if !allow {
				utils.Logger.Info(
					"max retry attempts reached, terminating",
					zap.String("address", r.RemoteAddr),
					zap.String("path", r.URL.Path),
				)
				http.Error(w, "Service not available", http.StatusServiceUnavailable)
				return
			}

			logger.Info(
				"Attempting retry",
				zap.String("address", r.RemoteAddr),
				zap.String("path", r.URL.Path),
				zap.Int("reties", attempts),
			)

			lb.Serve(
				w,
				r.WithContext(context.WithValue(r.Context(), loadbalancer.RETRIES_ATTEMPTED, attempts+1)),
			)

		}
		sp.AddBackend(backendServer)
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: http.HandlerFunc(lb.Serve),
	}

	go lb.LaunchHealthCheck(ctx, config.HealthCheckIntervalInSec)

	go func() {
		<-ctx.Done()
		shutdownCtx, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Fatal("Error in shutdown", zap.Error(err))
		}
	}()

	logger.Info(
		"load balancer started",
		zap.String("strategy", config.Strategy),
		zap.Int("port", config.Port),
	)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logger.Fatal("ListenAndServe() error", zap.Error(err))
	}
}
