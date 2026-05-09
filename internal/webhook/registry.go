package webhook

import (
	"strings"
	"sync"

	"github.com/google/uuid"
)

// Registry manages registered webhooks.
type Registry struct {
	mu       sync.RWMutex
	webhooks map[string]*Webhook
}

// NewRegistry creates a new webhook registry.
func NewRegistry() *Registry {
	return &Registry{
		webhooks: make(map[string]*Webhook),
	}
}

// Add registers a new webhook URL and returns its assigned ID.
func (r *Registry) Add(url string) *Webhook {
	r.mu.Lock()
	defer r.mu.Unlock()

	idStr := strings.ReplaceAll(uuid.New().String(), "-", "")
	if len(idStr) > 8 {
		idStr = idStr[:8]
	}
	whID := "wh-" + idStr

	wh := &Webhook{
		ID:  whID,
		URL: url,
	}
	r.webhooks[whID] = wh
	return wh
}

// GetAll returns a list of all registered webhooks.
func (r *Registry) GetAll() []*Webhook {
	r.mu.RLock()
	defer r.mu.RUnlock()

	all := make([]*Webhook, 0, len(r.webhooks))
	for _, wh := range r.webhooks {
		all = append(all, wh)
	}
	return all
}
