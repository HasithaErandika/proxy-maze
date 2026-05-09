package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// MetricsTracker interface to avoid cyclic dependency with internal/metrics
type MetricsTracker interface {
	IncrementWebhookDeliveries()
}

// Dispatcher handles the delivery of webhook payloads.
type Dispatcher struct {
	registry *Registry
	client   *http.Client
	metrics  MetricsTracker
}

// NewDispatcher creates a new webhook dispatcher.
func NewDispatcher(registry *Registry, metrics MetricsTracker) *Dispatcher {
	return &Dispatcher{
		registry: registry,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		metrics: metrics,
	}
}

// SetMetrics updates the metrics tracker.
func (d *Dispatcher) SetMetrics(metrics MetricsTracker) {
	d.metrics = metrics
}

// SendFired dispatches the alert.fired event to all webhooks.
func (d *Dispatcher) SendFired(alert interface{}) {
	d.dispatchAll("alert.fired", alert)
}

// SendResolved dispatches the alert.resolved event to all webhooks.
func (d *Dispatcher) SendResolved(alert interface{}) {
	d.dispatchAll("alert.resolved", alert)
}

func (d *Dispatcher) dispatchAll(event string, alert interface{}) {
	payload := Payload{
		Event: event,
		Alert: alert,
	}

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
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	backoff := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second, 16 * time.Second}
	maxRetries := len(backoff)

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(backoff[attempt-1]):
			case <-ctx.Done():
				return
			}
		}

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := d.client.Do(req)
		if err != nil {
			// Network error, retry
			continue
		}
		
		status := resp.StatusCode
		resp.Body.Close()

		if status >= 200 && status < 300 {
			// Success
			if d.metrics != nil {
				d.metrics.IncrementWebhookDeliveries()
			}
			return
		} else if status == 500 || status == 502 || status == 503 || status == 504 {
			// Retryable errors
			continue
		} else {
			// Non-retryable error (e.g. 400, 401, 404)
			return
		}
	}
}
