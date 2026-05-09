package proxy

import "time"

// Status constants
const (
	StatusPending = "pending"
	StatusUp      = "up"
	StatusDown    = "down"
)

// Proxy represents a tracked proxy server.
type Proxy struct {
	ID                  string        `json:"id"`
	URL                 string        `json:"url"`
	Status              string        `json:"status"`
	LastCheckedAt       *time.Time    `json:"last_checked_at"`
	ConsecutiveFailures int           `json:"consecutive_failures"`
	TotalChecks         int           `json:"total_checks"`
	UptimePercentage    float64       `json:"uptime_percentage"`
	History             []CheckRecord `json:"history"` // capped at 100 entries
}

// CheckRecord represents a single health check probe result.
type CheckRecord struct {
	CheckedAt time.Time `json:"checked_at"`
	Status    string    `json:"status"`
}
