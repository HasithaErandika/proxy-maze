package alert

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/HasithaErandika/proxy-maze/internal/webhook"
	"github.com/google/uuid"
)

type Manager struct {
	mu           sync.RWMutex
	alerts       []*Alert
	activeAlert  *Alert
	dispatcher   *webhook.Dispatcher
	igRegistry   IntegrationDispatcher 
	Threshold    float64
}

type IntegrationDispatcher interface {
	DispatchAlert(event string, a *Alert, dispatcher *webhook.Dispatcher)
}

func NewManager(dispatcher *webhook.Dispatcher) *Manager {
	return &Manager{
		alerts:     make([]*Alert, 0),
		dispatcher: dispatcher,
		Threshold:  0.20,
	}
}

func (m *Manager) SetIntegrationRegistry(ig IntegrationDispatcher) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.igRegistry = ig
}

func (m *Manager) Evaluate(totalProxies, failedProxies int, failedProxyIDs []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var failureRate float64
	if totalProxies > 0 {
		failureRate = float64(failedProxies) / float64(totalProxies)
	}

	isBreaching := failureRate >= m.Threshold

	if isBreaching && m.activeAlert == nil {
		idStr := strings.ReplaceAll(uuid.New().String(), "-", "")
		if len(idStr) > 6 {
			idStr = idStr[:6]
		}
		alertID := "alert-" + idStr

		newAlert := &Alert{
			AlertID:        alertID,
			Status:         StatusActive,
			FailureRate:    failureRate,
			TotalProxies:   totalProxies,
			FailedProxies:  failedProxies,
			FailedProxyIDs: failedProxyIDs,
			Threshold:      m.Threshold,
			FiredAt:        time.Now().UTC(),
			Message:        fmt.Sprintf("Proxy pool failure rate exceeded threshold (%.2f%%)", failureRate*100),
		}

		m.activeAlert = newAlert
		m.alerts = append(m.alerts, newAlert)

		if m.dispatcher != nil {
			payload := map[string]interface{}{
				"event":            "alert.fired",
				"alert_id":         newAlert.AlertID,
				"fired_at":         newAlert.FiredAt.Format(time.RFC3339),
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
		m.activeAlert.FailureRate = failureRate
		m.activeAlert.TotalProxies = totalProxies
		m.activeAlert.FailedProxies = failedProxies
		m.activeAlert.FailedProxyIDs = failedProxyIDs
	} else if !isBreaching && m.activeAlert != nil {
		now := time.Now().UTC()
		m.activeAlert.Status = StatusResolved
		m.activeAlert.ResolvedAt = &now
		m.activeAlert.FailureRate = failureRate
		m.activeAlert.TotalProxies = totalProxies
		m.activeAlert.FailedProxies = failedProxies
		m.activeAlert.FailedProxyIDs = failedProxyIDs
		
		resolvedAlert := m.activeAlert
		m.activeAlert = nil

		if m.dispatcher != nil {
			resolvedAtStr := ""
			if resolvedAlert.ResolvedAt != nil {
				resolvedAtStr = resolvedAlert.ResolvedAt.Format(time.RFC3339)
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
}

func (m *Manager) GetAll() []*Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshot := make([]*Alert, len(m.alerts))
	copy(snapshot, m.alerts)
	return snapshot
}

func (m *Manager) ActiveAlertCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.activeAlert != nil {
		return 1
	}
	return 0
}

func (m *Manager) TotalAlerts() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.alerts)
}
