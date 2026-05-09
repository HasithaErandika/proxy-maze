package alert

import "time"

const (
	StatusActive   = "active"
	StatusResolved = "resolved"
)

type Alert struct {
	AlertID        string     `json:"alert_id"`
	Status         string     `json:"status"`
	FailureRate    float64    `json:"failure_rate"`
	TotalProxies   int        `json:"total_proxies"`
	FailedProxies  int        `json:"failed_proxies"`
	FailedProxyIDs []string   `json:"failed_proxy_ids"`
	Threshold      float64    `json:"threshold"`
	FiredAt        time.Time  `json:"fired_at"`
	ResolvedAt     *time.Time `json:"resolved_at"`
	Message        string     `json:"message"`
}
