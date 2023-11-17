package serverpool

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"

	"github.com/abhikvarma/load-balancer/backend"
	"github.com/abhikvarma/load-balancer/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetNextValidPeerRoundRobin(t *testing.T) {
	mockServer1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer mockServer1.Close()
	dummyBackend1Url, err := url.Parse(mockServer1.URL)
	if err != nil {
		t.Fatal(err)
	}
	rp1 := httputil.NewSingleHostReverseProxy(dummyBackend1Url)
	dummyBackend1 := backend.NewBackend(dummyBackend1Url, rp1)

	mockServer2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer mockServer2.Close()
	dummyBackend2Url, err := url.Parse(mockServer2.URL)
	if err != nil {
		t.Fatal(err)
	}
	rp2 := httputil.NewSingleHostReverseProxy(dummyBackend2Url)
	dummyBackend2 := backend.NewBackend(dummyBackend2Url, rp2)

	sp, err := NewServerPool(utils.RoundRobin)
	if err != nil {
		t.Fatal(err)
	}

	sp.AddBackend(dummyBackend1)
	sp.AddBackend(dummyBackend2)

	peer1 := sp.GetNextValidPeer()
	peer2 := sp.GetNextValidPeer()
	peer3 := sp.GetNextValidPeer()

	assert.NotEqual(t, peer1, peer2)
	assert.Equal(t, peer1, peer3)
}
