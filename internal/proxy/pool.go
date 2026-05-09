package proxy

import (
	"strings"
	"sync"
)

type Pool struct {
	mu      sync.RWMutex
	proxies map[string]*Proxy
}

func NewPool() *Pool {
	return &Pool{
		proxies: make(map[string]*Proxy),
	}
}

func (p *Pool) Add(urls []string, replace bool) []*Proxy {
	p.mu.Lock()
	defer p.mu.Unlock()

	if replace {
		p.proxies = make(map[string]*Proxy)
	}

	var added []*Proxy
	for _, rawURL := range urls {
		id := extractIDFromURL(rawURL)
		
		if _, exists := p.proxies[id]; !exists {
			prx := &Proxy{
				ID:                  id,
				URL:                 rawURL,
				Status:              StatusPending,
				ConsecutiveFailures: 0,
				TotalChecks:         0,
				UptimePercentage:    0,
				History:             make([]CheckRecord, 0),
			}
			p.proxies[id] = prx
			added = append(added, prx)
		}
	}
	return added
}

func (p *Pool) Get(id string) *Proxy {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.proxies[id]
}

func (p *Pool) GetAll() []*Proxy {
	p.mu.RLock()
	defer p.mu.RUnlock()

	all := make([]*Proxy, 0, len(p.proxies))
	for _, prx := range p.proxies {
		all = append(all, prx)
	}
	return all
}

func (p *Pool) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.proxies = make(map[string]*Proxy)
}

func (p *Pool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.proxies)
}

func extractIDFromURL(rawURL string) string {
	parts := strings.Split(strings.TrimRight(rawURL, "/"), "/")
	if len(parts) > 0 {
		last := parts[len(parts)-1]
		if last != "" && !strings.Contains(last, ":") {
			return last
		}
	}
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "unknown"
}
