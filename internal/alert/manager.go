package alert

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/HasithaErandika/proxy-maze/internal/webhook"
)

const isoFormat = "2006-01-02T15:04:05Z"

// Manager handles the state machine for proxy alerts.
type Manager struct {
	mu           sync.RWMutex
	alerts       []*Alert
	activeAlert  *Alert
	dispatcher   *webhook.Dispatcher
	igRegistry   IntegrationDispatcher // using interface to prevent cycle
	Threshold    float64
}

// IntegrationDispatcher interface
type IntegrationDispatcher interface {
	DispatchAlert(event string, a *Alert, dispatcher *webhook.Dispatcher)
}

// NewManager creates a new alert manager.
func NewManager(dispatcher *webhook.Dispatcher) *Manager {
	return &Manager{
		alerts:     make([]*Alert, 0),
		dispatcher: dispatcher,
		Threshold:  0.20,
	}
}

// SetIntegrationRegistry sets the integration registry to dispatch to Slack/Discord.
func (m *Manager) SetIntegrationRegistry(ig IntegrationDispatcher) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.igRegistry = ig
}

// Evaluate checks the current pool state against the threshold and transitions the state machine.
func (m *Manager) Evaluate(totalProxies, failedProxies int, failedProxyIDs []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var failureRate float64
	if totalProxies > 0 {
		failureRate = float64(failedProxies) / float64(totalProxies)
	}

	isBreaching := failureRate >= m.Threshold

	if isBreaching && m.activeAlert == nil {
		// Fire new alert
		idStr := strings.ReplaceAll(uuid.New().String(), "-", "")
		if len(idStr) > 6 {
			idStr = idStr[:6]
		}
		alertID := "alert-" + idStr

		if failedProxyIDs == nil {
			failedProxyIDs = []string{}
		}

		newAlert := &Alert{
			AlertID:        alertID,
			Status:         StatusActive,
			FailureRate:    failureRate,
			TotalProxies:   totalProxies,
			FailedProxies:  failedProxies,
			FailedProxyIDs: failedProxyIDs,
			Threshold:      m.Threshold,
			FiredAt:        time.Now().UTC(),
			Message:        fmt.Sprintf("Proxy pool failure rate exceeded threshold"),
		}

		m.activeAlert = newAlert
		m.alerts = append(m.alerts, newAlert)

		// Dispatch async
		if m.dispatcher != nil {
			payload := map[string]interface{}{
				"event":            "alert.fired",
				"alert_id":         newAlert.AlertID,
				"fired_at":         newAlert.FiredAt.UTC().Format(isoFormat),
				"failure_rate":     newAlert.FailureRate,
				"total_proxies":    newAlert.TotalProxies,
				"failed_proxies":   newAlert.FailedProxies,
				"failed_proxy_ids": newAlert.FailedProxyIDs,
				"threshold":        newAlert.Threshold,
				"message":          newAlert.Message,
			}
			go m.dispatcher.SendFired(payload)
		}
		if m.igRegistry != nil {
			go m.igRegistry.DispatchAlert("alert.fired", newAlert, m.dispatcher)
		}

	} else if isBreaching && m.activeAlert != nil {
		// Update active alert with latest stats (optional but keeps it consistent)
		if failedProxyIDs == nil {
			failedProxyIDs = []string{}
		}
		m.activeAlert.FailureRate = failureRate
		m.activeAlert.TotalProxies = totalProxies
		m.activeAlert.FailedProxies = failedProxies
		m.activeAlert.FailedProxyIDs = failedProxyIDs
		// Do not dispatch again
	} else if !isBreaching && m.activeAlert != nil {
		// Resolve alert
		now := time.Now().UTC()
		m.activeAlert.Status = StatusResolved
		m.activeAlert.ResolvedAt = &now
		m.activeAlert.FailureRate = failureRate
		m.activeAlert.TotalProxies = totalProxies
		m.activeAlert.FailedProxies = failedProxies
		if failedProxyIDs == nil {
			failedProxyIDs = []string{}
		}
		m.activeAlert.FailedProxyIDs = failedProxyIDs
		
		resolvedAlert := m.activeAlert
		m.activeAlert = nil

		// Dispatch async
		if m.dispatcher != nil {
			resolvedAtStr := ""
			if resolvedAlert.ResolvedAt != nil {
				resolvedAtStr = resolvedAlert.ResolvedAt.UTC().Format(isoFormat)
			}
			payload := map[string]interface{}{
				"event":       "alert.resolved",
				"alert_id":    resolvedAlert.AlertID,
				"resolved_at": resolvedAtStr,
			}
			go m.dispatcher.SendResolved(payload)
		}
		if m.igRegistry != nil {
			go m.igRegistry.DispatchAlert("alert.resolved", resolvedAlert, m.dispatcher)
		}
	}
	// If !isBreaching && m.activeAlert == nil -> do nothing
}

// GetAll returns a copy of the alerts archive.
func (m *Manager) GetAll() []*Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a shallow copy of the slice to avoid concurrent modification panics
	snapshot := make([]*Alert, len(m.alerts))
	copy(snapshot, m.alerts)
	return snapshot
}

// ActiveAlertCount returns the number of currently active alerts (0 or 1).
func (m *Manager) ActiveAlertCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.activeAlert != nil {
		return 1
	}
	return 0
}

// TotalAlerts returns the total number of alerts in the archive.
func (m *Manager) TotalAlerts() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.alerts)
}
