package integration

import (
	"sync"
	"github.com/HasithaErandika/proxy-maze/internal/alert"
	"github.com/HasithaErandika/proxy-maze/internal/webhook"
)

// Registry manages external integrations and dispatches events.
type Registry struct {
	mu           sync.RWMutex
	integrations []Integration
}

// NewRegistry creates a new integration registry.
func NewRegistry() *Registry {
	return &Registry{
		integrations: make([]Integration, 0),
	}
}

// Add registers a new integration.
func (r *Registry) Add(integration Integration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.integrations = append(r.integrations, integration)
}

// DispatchAlert handles formatting and sending alerts to configured integrations.
func (r *Registry) DispatchAlert(event string, a *alert.Alert, dispatcher *webhook.Dispatcher) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, ig := range r.integrations {
		// Check if integration listens to this event
		listens := false
		for _, e := range ig.Events {
			if e == event {
				listens = true
				break
			}
		}
		if !listens {
			continue
		}

		if ig.Type == "slack" {
			go sendSlack(ig, event, a)
		} else if ig.Type == "discord" {
			go sendDiscord(ig, event, a)
		}
	}
}
