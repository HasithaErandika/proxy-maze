package config

import (
	"context"
	"sync"
)

// Store holds thread-safe runtime configurations.
type Store struct {
	mu                   sync.RWMutex
	checkIntervalSeconds int
	requestTimeoutMs     int
	cancelTicker         context.CancelFunc
}

// NewStore creates a new config store with default values.
func NewStore() *Store {
	return &Store{
		checkIntervalSeconds: 30,
		requestTimeoutMs:     5000,
	}
}

// Get returns the current configuration values.
func (s *Store) Get() (intervalSecs int, timeoutMs int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.checkIntervalSeconds, s.requestTimeoutMs
}

// Update updates the configuration and triggers the ticker cancellation if provided.
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

// SetTickerCancel sets the cancellation function for the background ticker loop.
// This allows the config store to restart the loop when config changes.
func (s *Store) SetTickerCancel(cancel context.CancelFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cancelTicker = cancel
}
