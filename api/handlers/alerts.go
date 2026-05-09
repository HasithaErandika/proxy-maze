package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/HasithaErandika/proxy-maze/internal/alert"
)

// AlertsHandler handles GET /alerts
func AlertsHandler(manager *alert.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		alerts := manager.GetAll()

		// Build response with explicit timestamp formatting
		type alertView struct {
			AlertID        string   `json:"alert_id"`
			Status         string   `json:"status"`
			FailureRate    float64  `json:"failure_rate"`
			TotalProxies   int      `json:"total_proxies"`
			FailedProxies  int      `json:"failed_proxies"`
			FailedProxyIDs []string `json:"failed_proxy_ids"`
			Threshold      float64  `json:"threshold"`
			FiredAt        string   `json:"fired_at"`
			ResolvedAt     *string  `json:"resolved_at"`
			Message        string   `json:"message"`
		}

		views := make([]alertView, 0, len(alerts))
		for _, a := range alerts {
			var resolvedAt *string
			if a.ResolvedAt != nil {
				s := a.ResolvedAt.UTC().Format("2006-01-02T15:04:05Z")
				resolvedAt = &s
			}

			failedIDs := a.FailedProxyIDs
			if failedIDs == nil {
				failedIDs = []string{}
			}

			views = append(views, alertView{
				AlertID:        a.AlertID,
				Status:         a.Status,
				FailureRate:    a.FailureRate,
				TotalProxies:   a.TotalProxies,
				FailedProxies:  a.FailedProxies,
				FailedProxyIDs: failedIDs,
				Threshold:      a.Threshold,
				FiredAt:        a.FiredAt.UTC().Format("2006-01-02T15:04:05Z"),
				ResolvedAt:     resolvedAt,
				Message:        a.Message,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(views)
	}
}
