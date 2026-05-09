package metrics

import (
	"sync/atomic"

	"github.com/HasithaErandika/proxy-maze/internal/alert"
	"github.com/HasithaErandika/proxy-maze/internal/proxy"
)

type Metrics struct {
	TotalChecks       int64
	WebhookDeliveries int64
	
	poolManager  *proxy.Pool
	alertManager *alert.Manager
}

func NewMetrics(poolManager *proxy.Pool, alertManager *alert.Manager) *Metrics {
	return &Metrics{
		TotalChecks:       0,
		WebhookDeliveries: 0,
		poolManager:       poolManager,
		alertManager:      alertManager,
	}
}

func (m *Metrics) IncrementTotalChecks() {
	atomic.AddInt64(&m.TotalChecks, 1)
}

func (m *Metrics) IncrementWebhookDeliveries() {
	atomic.AddInt64(&m.WebhookDeliveries, 1)
}

func (m *Metrics) GetSnapshot() map[string]interface{} {
	poolSize := 0
	if m.poolManager != nil {
		poolSize = m.poolManager.Size()
	}

	activeAlerts := 0
	totalAlerts := 0
	if m.alertManager != nil {
		activeAlerts = m.alertManager.ActiveAlertCount()
		totalAlerts = m.alertManager.TotalAlerts()
	}

	return map[string]interface{}{
		"total_checks":       atomic.LoadInt64(&m.TotalChecks),
		"current_pool_size":  poolSize,
		"active_alerts":      activeAlerts,
		"total_alerts":       totalAlerts,
		"webhook_deliveries": atomic.LoadInt64(&m.WebhookDeliveries),
	}
}
