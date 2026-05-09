package proxy

import (
	"context"
	"net/http"
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

// Start begins the background check loop. It will restart if ctx is cancelled.
func (c *Checker) Start(ctx context.Context) {
	for {
		intervalSecs, timeoutMs := c.config.Get()
		interval := time.Duration(intervalSecs) * time.Second

		c.client.Timeout = time.Duration(timeoutMs) * time.Millisecond

		ticker := time.NewTicker(interval)

		select {
		case <-ctx.Done():
			ticker.Stop()
			return // graceful shutdown
		default:
			c.runLoop(ctx, ticker)
		}
	}
}

func (c *Checker) runLoop(ctx context.Context, ticker *time.Ticker) {
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// The config was updated (or app is shutting down)
			// Context canceled, break inner loop to pick up new config
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

	downCount := 0
	var failedIDs []string

	for _, prx := range proxies {
		status := c.checkProxy(prx.URL)
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

		// Calculate Uptime
		upChecks := 0
		if prx.TotalChecks > 0 {
			for _, rec := range prx.History {
				if rec.Status == StatusUp {
					upChecks++
				}
			}
			if status == StatusUp {
				upChecks++
			}
			prx.UptimePercentage = float64(upChecks) / float64(prx.TotalChecks)
		}

		// Append History
		record := CheckRecord{
			CheckedAt: now,
			Status:    prx.Status,
		}
		prx.History = append(prx.History, record)
		if len(prx.History) > 100 {
			prx.History = prx.History[1:] // keep last 100
		}

		if prx.Status == StatusDown {
			downCount++
			failedIDs = append(failedIDs, prx.ID)
		}
		c.pool.mu.Unlock()
	}

	c.alertM.Evaluate(len(proxies), downCount, failedIDs)
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
