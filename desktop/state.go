package desktop

import (
	"grout/internal"
	"grout/romm"
	"sync"
)

type AppState struct {
	mu     sync.RWMutex
	config *internal.Config
	host   *romm.Host
	
	// Listeners for state changes
	listeners []func()
}

func NewAppState() *AppState {
	cfg, err := internal.LoadConfig()
	if err != nil {
		cfg = &internal.Config{}
	}

	var host *romm.Host
	if cfg != nil && len(cfg.Hosts) > 0 {
		host = &cfg.Hosts[0]
	}
	
	return &AppState{
		config: cfg,
		host:   host,
	}
}

func (s *AppState) GetConfig() *internal.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

func (s *AppState) GetHost() *romm.Host {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.host
}

func (s *AppState) SetHost(h *romm.Host) {
	s.mu.Lock()
	s.host = h
	s.mu.Unlock()
	s.notify()
}

func (s *AppState) AddListener(l func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listeners = append(s.listeners, l)
}

func (s *AppState) notify() {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, l := range s.listeners {
		l()
	}
}
