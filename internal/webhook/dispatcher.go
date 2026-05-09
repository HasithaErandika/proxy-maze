package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type MetricsTracker interface {
	IncrementWebhookDeliveries()
}

type Dispatcher struct {
	registry *Registry
	client   *http.Client
	metrics  MetricsTracker
}

func NewDispatcher(registry *Registry, metrics MetricsTracker) *Dispatcher {
	return &Dispatcher{
		registry: registry,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		metrics: metrics,
	}
}

func (d *Dispatcher) SetMetrics(metrics MetricsTracker) {
	d.metrics = metrics
}

func (d *Dispatcher) SendFired(payload interface{}) {
	d.dispatchAll(payload)
}

func (d *Dispatcher) SendResolved(payload interface{}) {
	d.dispatchAll(payload)
}

func (d *Dispatcher) dispatchAll(payload interface{}) {
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}

	webhooks := d.registry.GetAll()
	for _, wh := range webhooks {
		go d.deliverWithRetry(wh.URL, body)
	}
}

func (d *Dispatcher) deliverWithRetry(url string, body []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	defer cancel()

	success := DeliverWithRetry(ctx, d.client, url, body)
	if success && d.metrics != nil {
		d.metrics.IncrementWebhookDeliveries()
	}
}

func DeliverWithRetry(ctx context.Context, client *http.Client, url string, body []byte) bool {
	backoff := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second, 16 * time.Second}
	maxRetries := len(backoff)

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(backoff[attempt-1]):
			case <-ctx.Done():
				return false
			}
		}

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		
		status := resp.StatusCode
		resp.Body.Close()

		if status >= 200 && status < 300 {
			return true
		} else if status == 500 || status == 502 || status == 503 || status == 504 {
			continue
		} else {
			return false
		}
	}
	return false
}
