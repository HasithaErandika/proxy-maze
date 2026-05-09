package proxy

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/HasithaErandika/proxy-maze/internal/alert"
	"github.com/HasithaErandika/proxy-maze/internal/config"
)

type MetricsTracker interface {
	IncrementTotalChecks()
}

type Checker struct {
	pool    *Pool
	config  *config.Store
	alertM  *alert.Manager
	metrics MetricsTracker
	client  *http.Client
}

func NewChecker(pool *Pool, cfg *config.Store, alertM *alert.Manager, metrics MetricsTracker) *Checker {
	return &Checker{
		pool:    pool,
		config:  cfg,
		alertM:  alertM,
		metrics: metrics,
		client:  &http.Client{},
	}
}

func (c *Checker) Start(ctx context.Context) {
	intervalSecs, _ := c.config.Get()
	interval := time.Duration(intervalSecs) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.executeChecks(ctx)

			newIntervalSecs, _ := c.config.Get()
			newInterval := time.Duration(newIntervalSecs) * time.Second
			if newInterval != interval {
				interval = newInterval
				ticker.Reset(interval)
			}
		}
	}
}

func (c *Checker) executeChecks(ctx context.Context) {
	proxies := c.pool.GetAll()
	if len(proxies) == 0 {
		return 
	}

	_, timeoutMs := c.config.Get()
	timeout := time.Duration(timeoutMs) * time.Millisecond

	type result struct {
		prx    *Proxy
		status string
		now    time.Time
	}
	results := make([]result, len(proxies))

	var wg sync.WaitGroup
	for i, prx := range proxies {
		wg.Add(1)
		go func(i int, p *Proxy) {
			defer wg.Done()
			reqCtx, reqCancel := context.WithTimeout(ctx, timeout)
			defer reqCancel()
			status := c.checkProxy(reqCtx, p.URL)
			results[i] = result{p, status, time.Now().UTC()}
		}(i, prx)
	}
	wg.Wait()

	if ctx.Err() != nil {
		return 
	}

	downCount := 0
	var failedIDs []string

	c.pool.mu.Lock()
	for _, res := range results {
		if _, exists := c.pool.proxies[res.prx.ID]; !exists {
			continue
		}
		prx := res.prx
		status := res.status
		now := res.now

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

		record := CheckRecord{
			CheckedAt: now,
			Status:    prx.Status,
		}
		prx.History = append(prx.History, record)
		if len(prx.History) > 100 {
			prx.History = prx.History[1:] 
		}

		if prx.Status == StatusDown {
			downCount++
			failedIDs = append(failedIDs, prx.ID)
		}
	}
	currentSize := len(c.pool.proxies)
	c.pool.mu.Unlock()

	c.alertM.Evaluate(currentSize, downCount, failedIDs)
}

func (c *Checker) checkProxy(ctx context.Context, urlStr string) string {
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
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
