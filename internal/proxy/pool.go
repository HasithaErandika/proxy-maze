package proxy

import (
	"strings"
	"sync"
)

// Pool manages a thread-safe map of tracked proxies.
type Pool struct {
	mu      sync.RWMutex
	proxies map[string]*Proxy
}

// NewPool creates a new proxy pool.
func NewPool() *Pool {
	return &Pool{
		proxies: make(map[string]*Proxy),
	}
}

// Add adds multiple URLs to the pool.
// If replace is true, the pool is cleared before adding.
func (p *Pool) Add(urls []string, replace bool) []*Proxy {
	p.mu.Lock()
	defer p.mu.Unlock()

	if replace {
		p.proxies = make(map[string]*Proxy)
	}

	var added []*Proxy
	for _, rawURL := range urls {
		id := extractIDFromURL(rawURL)
		
		// If already exists, we skip adding it again to avoid resetting its state,
		// unless replace was true, in which case it wouldn't exist anyway.
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

// Get returns a single proxy by ID, or nil if not found.
// The caller should ideally not mutate the returned Proxy without lock, 
// but since Go maps return pointers, mutations should be coordinated via the pool.
// For read-only API responses, we can return a copy if strict safety is needed,
// but for simplicity we return the pointer. We will protect writes inside Checker.
func (p *Pool) Get(id string) *Proxy {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.proxies[id]
}

// GetAll returns a slice of all proxies currently in the pool.
func (p *Pool) GetAll() []*Proxy {
	p.mu.RLock()
	defer p.mu.RUnlock()

	all := make([]*Proxy, 0, len(p.proxies))
	for _, prx := range p.proxies {
		all = append(all, prx)
	}
	return all
}

// Clear removes all proxies from the pool.
func (p *Pool) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.proxies = make(map[string]*Proxy)
}

// Size returns the total number of proxies in the pool.
func (p *Pool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.proxies)
}

// extractIDFromURL extracts the last path segment of the URL to use as ID.
// Fallback to the full URL if no path is found.
func extractIDFromURL(rawURL string) string {
	parts := strings.Split(strings.TrimRight(rawURL, "/"), "/")
	if len(parts) > 0 {
		last := parts[len(parts)-1]
		if last != "" && !strings.Contains(last, ":") { // naive check to avoid picking up the host port as ID if no path
			return last
		}
	}
	// Fallback mechanism to generate a safe ID if extraction fails or is unhelpful
	// But according to prompt: "Extract ID from last path segment of URL"
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "unknown"
}
