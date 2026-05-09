package webhook

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type MetricsTracker interface {
	IncrementWebhookDeliveries()
}

type Job struct {
	URL       string
	Body      []byte
	Key       string
	CreatedAt time.Time
	NextRun   time.Time
	Attempts  int
	InFlight  bool
}

type Dispatcher struct {
	registry    *Registry
	client      *http.Client
	metrics     MetricsTracker
	mu          sync.Mutex
	queue       map[string]*Job
	successKeys map[string]struct{}
	started     bool
}

func NewDispatcher(registry *Registry, metrics MetricsTracker) *Dispatcher {
	return &Dispatcher{
		registry:    registry,
		client:      &http.Client{},
		metrics:     metrics,
		queue:       make(map[string]*Job),
		successKeys: make(map[string]struct{}),
	}
}

func (d *Dispatcher) SetMetrics(metrics MetricsTracker) {
	d.mu.Lock()
	defer d.mu.Unlock()
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
		d.EnqueueDelivery(wh.URL, body)
	}
}

func computeDeliveryKey(url string, payload []byte) string {
	var p map[string]interface{}
	json.Unmarshal(payload, &p)
	
	event := ""
	if e, ok := p["event"].(string); ok {
		event = e
	}
	
	alertId := ""
	if a, ok := p["alert_id"].(string); ok {
		alertId = a
	}

	if alertId == "" {
		hash := sha1.Sum(payload)
		alertId = fmt.Sprintf("%x", hash)
	}
	return fmt.Sprintf("%s|%s|%s", url, event, alertId)
}

func (d *Dispatcher) EnqueueDelivery(url string, body []byte) {
	key := computeDeliveryKey(url, body)

	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.successKeys[key]; ok {
		return
	}
	if _, ok := d.queue[key]; ok {
		return
	}

	d.queue[key] = &Job{
		URL:       url,
		Body:      body,
		Key:       key,
		CreatedAt: time.Now(),
		NextRun:   time.Now(),
	}

	if !d.started {
		d.started = true
		go d.processLoop()
	}
}

func (d *Dispatcher) processLoop() {
	ticker := time.NewTicker(250 * time.Millisecond)
	for range ticker.C {
		d.mu.Lock()

		now := time.Now()
		var pending []*Job
		for key, job := range d.queue {
			if job.InFlight {
				continue
			}
			if now.Before(job.NextRun) {
				continue
			}
			if time.Since(job.CreatedAt) > 60*time.Second {
				delete(d.queue, key)
				continue
			}

			pending = append(pending, job)
			if len(pending) >= 25 {
				break
			}
		}

		for _, job := range pending {
			job.InFlight = true
		}
		d.mu.Unlock()

		for _, job := range pending {
			go d.attemptJob(job)
		}
	}
}

func (d *Dispatcher) attemptJob(job *Job) {
	reqCtx, reqCancel := context.WithTimeout(context.Background(), 5*time.Second)
	req, err := http.NewRequestWithContext(reqCtx, "POST", job.URL, bytes.NewReader(job.Body))
	if err != nil {
		reqCancel()
		d.requeue(job, false)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		reqCancel()
		d.requeue(job, true)
		return
	}

	status := resp.StatusCode
	resp.Body.Close()
	reqCancel()

	if status >= 200 && status < 300 {
		d.markSuccess(job)
	} else if status >= 500 && status < 600 {
		d.requeue(job, true)
	} else {
		d.requeue(job, false)
	}
}

func (d *Dispatcher) requeue(job *Job, transient bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !transient {
		delete(d.queue, job.Key)
		return
	}

	j, ok := d.queue[job.Key]
	if !ok {
		return
	}

	j.Attempts++
	j.NextRun = time.Now().Add(1500 * time.Millisecond)
	j.InFlight = false
}

func (d *Dispatcher) markSuccess(job *Job) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.successKeys[job.Key]; !ok {
		d.successKeys[job.Key] = struct{}{}
		if d.metrics != nil {
			d.metrics.IncrementWebhookDeliveries()
		}
	}
	delete(d.queue, job.Key)
}
