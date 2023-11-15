package serverpool

import (
	"github.com/abhikvarma/load-balancer/backend"
)

type lcServerPool struct {
	backends []backend.Backend
}

func (s *lcServerPool) GetNextValidPeer() backend.Backend {
	var leastConnectedPeer backend.Backend
	for _, b := range s.backends {
		if b.IsAlive() {
			leastConnectedPeer = b
			break
		}
	}

	for _, b := range s.backends {
		if !b.IsAlive() {
			continue
		}
		if leastConnectedPeer.GetActiveConnections() > b.GetActiveConnections() {
			leastConnectedPeer = b
		}
	}
	return leastConnectedPeer
}

func (s *lcServerPool) GetBackends() []backend.Backend {
	return s.backends
}

func (s *lcServerPool) AddBackend(b backend.Backend) {
	s.backends = append(s.backends, b)
}

func (s *lcServerPool) GetServerPoolSize() int {
	return len(s.backends)
}
