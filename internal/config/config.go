package config

import (
	"context"
	"sync"
)

type Store struct {
	mu                   sync.RWMutex
	checkIntervalSeconds int
	requestTimeoutMs     int
	cancelTicker         context.CancelFunc
}

func NewStore() *Store {
	return &Store{
		checkIntervalSeconds: 30,
		requestTimeoutMs:     5000,
	}
}

func (s *Store) Get() (intervalSecs int, timeoutMs int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.checkIntervalSeconds, s.requestTimeoutMs
}

func (s *Store) Update(intervalSecs int, timeoutMs int) {
	s.mu.Lock()
	s.checkIntervalSeconds = intervalSecs
	s.requestTimeoutMs = timeoutMs
	cancel := s.cancelTicker
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}
}

func (s *Store) SetTickerCancel(cancel context.CancelFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cancelTicker = cancel
}
