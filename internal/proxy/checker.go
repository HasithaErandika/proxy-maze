package proxy

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/HasithaErandika/proxy-maze/internal/alert"
	"github.com/HasithaErandika/proxy-maze/internal/config"
)

// MetricsTracker interface to avoid cyclic dependency
type MetricsTracker interface {
	IncrementTotalChecks()
}

// Checker handles the background polling of proxies.
type Checker struct {
	pool    *Pool
	config  *config.Store
	alertM  *alert.Manager
	metrics MetricsTracker
	client  *http.Client
}

// NewChecker creates a new background checker.
func NewChecker(pool *Pool, cfg *config.Store, alertM *alert.Manager, metrics MetricsTracker) *Checker {
	return &Checker{
		pool:    pool,
		config:  cfg,
		alertM:  alertM,
		metrics: metrics,
		client:  &http.Client{},
	}
}

// Start begins the background check loop. It runs until ctx is cancelled.
func (c *Checker) Start(ctx context.Context) {
	intervalSecs, timeoutMs := c.config.Get()
	interval := time.Duration(intervalSecs) * time.Second
	c.client.Timeout = time.Duration(timeoutMs) * time.Millisecond

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run an initial check immediately
	c.executeChecks()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.executeChecks()
		}
	}
}

func (c *Checker) executeChecks() {
	proxies := c.pool.GetAll()
	if len(proxies) == 0 {
		c.alertM.Evaluate(0, 0, nil)
		return
	}

	// Update timeout from config before each round
	_, timeoutMs := c.config.Get()
	c.client.Timeout = time.Duration(timeoutMs) * time.Millisecond

	downCount := 0
	var failedIDs []string

	// Check all proxies concurrently with a WaitGroup
	type result struct {
		prx    *Proxy
		status string
	}

	results := make([]result, len(proxies))
	var wg sync.WaitGroup

	for i, prx := range proxies {
		wg.Add(1)
		go func(idx int, p *Proxy) {
			defer wg.Done()
			status := c.checkProxy(p.URL)
			results[idx] = result{prx: p, status: status}
		}(i, prx)
	}
	wg.Wait()

	// Now update all proxies under lock
	for _, res := range results {
		prx := res.prx
		status := res.status
		now := time.Now().UTC()

		c.pool.mu.Lock()
		prx.LastCheckedAt = &now
		prx.TotalChecks++
		if c.metrics != nil {
			c.metrics.IncrementTotalChecks()
		}

		if status == StatusUp {
			prx.Status = StatusUp
			prx.ConsecutiveFailures = 0
		} else {
			prx.Status = StatusDown
			prx.ConsecutiveFailures++
		}

		// Append History record
		record := CheckRecord{
			CheckedAt: now,
			Status:    prx.Status,
		}
		prx.History = append(prx.History, record)
		if len(prx.History) > 100 {
			prx.History = prx.History[1:] // keep last 100
		}

		// Recalculate uptime from full history
		upChecks := 0
		for _, rec := range prx.History {
			if rec.Status == StatusUp {
				upChecks++
			}
		}
		if prx.TotalChecks > 0 {
			prx.UptimePercentage = float64(upChecks) / float64(prx.TotalChecks)
		}

		if prx.Status == StatusDown {
			downCount++
			failedIDs = append(failedIDs, prx.ID)
		}
		c.pool.mu.Unlock()
	}

	c.alertM.Evaluate(len(proxies), downCount, failedIDs)
	log.Printf("[Checker] Checked %d proxies: %d down, failure_rate=%.2f", len(proxies), downCount, float64(downCount)/float64(len(proxies)))
}

func (c *Checker) checkProxy(urlStr string) string {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return StatusDown
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return StatusDown
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return StatusUp
	}
	return StatusDown
}
