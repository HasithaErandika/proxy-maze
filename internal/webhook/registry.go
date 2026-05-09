package webhook

import (
	"strings"
	"sync"

	"github.com/google/uuid"
)

type Registry struct {
	mu       sync.RWMutex
	webhooks map[string]*Webhook
}

func NewRegistry() *Registry {
	return &Registry{
		webhooks: make(map[string]*Webhook),
	}
}

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

func (r *Registry) GetAll() []*Webhook {
	r.mu.RLock()
	defer r.mu.RUnlock()

	all := make([]*Webhook, 0, len(r.webhooks))
	for _, wh := range r.webhooks {
		all = append(all, wh)
	}
	return all
}
