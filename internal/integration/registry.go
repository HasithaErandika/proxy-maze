package integration

import (
	"sync"

	"github.com/HasithaErandika/proxy-maze/internal/alert"
	"github.com/HasithaErandika/proxy-maze/internal/webhook"
)

type Registry struct {
	mu           sync.RWMutex
	integrations []Integration
}

func NewRegistry() *Registry {
	return &Registry{
		integrations: make([]Integration, 0),
	}
}

func (r *Registry) Add(integration Integration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.integrations = append(r.integrations, integration)
}

func (r *Registry) DispatchAlert(event string, a *alert.Alert, dispatcher *webhook.Dispatcher) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, ig := range r.integrations {
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

		switch ig.Type {
		case "slack":
			go sendSlack(ig, event, a)
		case "discord":
			go sendDiscord(ig, event, a)
		}
	}
}
