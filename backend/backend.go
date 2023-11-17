package backend

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/charmbracelet/log"
)

type Backend interface {
	SetAlive(bool)
	IsAlive() bool
	GetURL() *url.URL
	GetActiveConnections() int
	Serve(http.ResponseWriter, *http.Request)
}

type backend struct {
	url          *url.URL
	alive        bool
	mux          sync.RWMutex
	connections  int
	reverseProxy *httputil.ReverseProxy
}

func (b *backend) GetActiveConnections() int {
	b.mux.RLock()
	connections := b.connections
	b.mux.RUnlock()
	return connections
}

func (b *backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.alive = alive
	b.mux.Unlock()
}

func (b *backend) IsAlive() bool {
	b.mux.RLock()
	alive := b.alive
	defer b.mux.RUnlock()
	return alive
}

func (b *backend) GetURL() *url.URL {
	return b.url
}

func (b *backend) Serve(w http.ResponseWriter, r *http.Request) {
	defer func() {
		b.mux.Lock()
		b.connections--
		b.mux.Unlock()
	}()

	b.mux.Lock()
	b.connections++
	log.Info(fmt.Sprintf("Sending request %s to %s", r.RequestURI, b.GetURL()))
	b.mux.Unlock()
	b.reverseProxy.ServeHTTP(w, r)
}

func IsBackendAlive(ctx context.Context, aliveChanel chan bool, u *url.URL) {
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", u.Host)
	if err != nil {
		log.Debug("Site Unreachable", err)
		aliveChanel <- false
		return
	}
	_ = conn.Close()
	aliveChanel <- true
}

func NewBackend(u *url.URL, rp *httputil.ReverseProxy) Backend {
	return &backend{
		url:          u,
		alive:        true,
		reverseProxy: rp,
	}
}
