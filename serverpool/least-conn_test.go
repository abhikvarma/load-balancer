package serverpool

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"
	"time"

	"github.com/abhikvarma/load-balancer/backend"
	"github.com/abhikvarma/load-balancer/utils"
	"github.com/stretchr/testify/assert"
)

func dummyBackend(w http.ResponseWriter, r *http.Request) {
	time.Sleep(3 * time.Second)
}

func TestGetNextValidPeer(t *testing.T) {
	mockServer1 := httptest.NewServer(http.HandlerFunc(dummyBackend))
	defer mockServer1.Close()
	dummyBackend1Url, err := url.Parse(mockServer1.URL)
	if err != nil {
		t.Fatal(err)
	}
	rp1 := httputil.NewSingleHostReverseProxy(dummyBackend1Url)
	dummyBackend1 := backend.NewBackend(dummyBackend1Url, rp1)

	mockServer2 := httptest.NewServer(http.HandlerFunc(dummyBackend))
	defer mockServer2.Close()
	dummyBackend2Url, err := url.Parse(mockServer2.URL)
	if err != nil {
		t.Fatal(err)
	}
	rp2 := httputil.NewSingleHostReverseProxy(dummyBackend2Url)
	dummyBackend2 := backend.NewBackend(dummyBackend2Url, rp2)

	sp, err := NewServerPool(utils.LeastConnected)
	if err != nil {
		t.Fatal(err)
	}

	sp.AddBackend(dummyBackend1)
	sp.AddBackend(dummyBackend2)

	done := make(chan struct{})

	peer1 := sp.GetNextValidPeer()
	assert.NotNil(t, peer1)
	go func() {
		defer close(done)
		peer1.Serve(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/test", nil))
	}()

	time.Sleep(1 * time.Second)
	peer2 := sp.GetNextValidPeer()
	assert.NotNil(t, peer2)

	assert.Equal(t, 0, peer2.GetActiveConnections())
	assert.NotEqual(t, peer1, peer2)

	<-done
}
