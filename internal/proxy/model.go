package proxy

import "time"

const (
	StatusPending = "pending"
	StatusUp      = "up"
	StatusDown    = "down"
)

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

type CheckRecord struct {
	CheckedAt time.Time `json:"checked_at"`
	Status    string    `json:"status"`
}

func (p *Proxy) Clone() *Proxy {
	clone := *p
	if p.LastCheckedAt != nil {
		t := *p.LastCheckedAt
		clone.LastCheckedAt = &t
	}
	if p.History != nil {
		clone.History = make([]CheckRecord, len(p.History))
		copy(clone.History, p.History)
	} else {
		clone.History = make([]CheckRecord, 0)
	}
	return &clone
}
