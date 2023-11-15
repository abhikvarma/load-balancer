package backend

import (
	"net/http/httputil"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

const urlString string = "http://localhost:8080"

func TestNewBackend(t *testing.T) {
	url, _ := url.Parse(urlString)
	b := NewBackend(url, httputil.NewSingleHostReverseProxy(url))
	assert.Equal(t, urlString, b.GetURL().String())
	assert.Equal(t, true, b.IsAlive())
}

func TestBackendAlive(t *testing.T) {
	url, _ := url.Parse(urlString)
	b := NewBackend(url, httputil.NewSingleHostReverseProxy(url))
	b.SetAlive(false)
	assert.Equal(t, false, b.IsAlive())
	b.SetAlive(true)
	assert.Equal(t, true, b.IsAlive())
}
