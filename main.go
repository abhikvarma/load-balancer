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
	"github.com/charmbracelet/log"
)

func main() {
	config, err := utils.GetLBConfig()
	if err != nil {
		log.Fatal(err.Error())
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sp, err := serverpool.NewServerPool(utils.GetLBStrategy(config.Strategy))
	if err != nil {
		log.Fatal(err.Error())
	}
	lb := loadbalancer.NewLoadBalancer(sp)

	for _, b := range config.Backends {
		endpoint, err := url.Parse(b)
		if err != nil {
			log.Fatal(err.Error())
		}

		rp := httputil.NewSingleHostReverseProxy(endpoint)
		backendServer := backend.NewBackend(endpoint, rp)
		rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
			log.Info(fmt.Sprintf("error handling the request %s", r.RequestURI), "host", endpoint.Host, e)

			attempts, allow := loadbalancer.AllowRetry(r, config.MaxRetries)
			if !allow {
				log.Info(
					"max retry attempts reached, terminating", "address", r.RemoteAddr, "path", r.URL.Path,
				)
				http.Error(w, "Service not available", http.StatusServiceUnavailable)
				return
			}

			log.Info(
				"Attempting retry", "address", r.RemoteAddr, "path", r.URL.Path, "reties", attempts,
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
			log.Fatal("Error in shutdown", err)
		}
	}()

	log.Info(
		"load balancer started", "strategy", config.Strategy, "port", config.Port,
	)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal("ListenAndServe() error", err)
	}
}
