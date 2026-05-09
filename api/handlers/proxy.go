package handlers

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/HasithaErandika/proxy-maze/internal/proxy"
)

// ProxyHandler handles GET /proxies/{id} and GET /proxies/{id}/history
func ProxyHandler(pool *proxy.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse ID and potential "history" suffix from path
		path := strings.TrimPrefix(r.URL.Path, "/proxies/")
		parts := strings.Split(path, "/")

		if len(parts) == 0 || parts[0] == "" {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		id := parts[0]
		isHistory := len(parts) > 1 && parts[1] == "history"

		prx := pool.Get(id)
		if prx == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Not Found"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if isHistory {
			// Return array of {checked_at, status}
			type historyEntry struct {
				CheckedAt string `json:"checked_at"`
				Status    string `json:"status"`
			}
			entries := make([]historyEntry, 0, len(prx.History))
			for _, h := range prx.History {
				entries = append(entries, historyEntry{
					CheckedAt: h.CheckedAt.UTC().Format("2006-01-02T15:04:05Z"),
					Status:    h.Status,
				})
			}
			json.NewEncoder(w).Encode(entries)
		} else {
			// Build the full dossier response with explicit timestamp formatting
			type historyEntry struct {
				CheckedAt string `json:"checked_at"`
				Status    string `json:"status"`
			}
			histEntries := make([]historyEntry, 0, len(prx.History))
			for _, h := range prx.History {
				histEntries = append(histEntries, historyEntry{
					CheckedAt: h.CheckedAt.UTC().Format("2006-01-02T15:04:05Z"),
					Status:    h.Status,
				})
			}

			var lastChecked *string
			if prx.LastCheckedAt != nil {
				s := prx.LastCheckedAt.UTC().Format("2006-01-02T15:04:05Z")
				lastChecked = &s
			}

			// uptime_percentage should be 0-100 range, rounded to 1 decimal
			uptimePct := math.Round(prx.UptimePercentage*1000) / 10

			resp := map[string]interface{}{
				"id":                    prx.ID,
				"url":                   prx.URL,
				"status":                prx.Status,
				"last_checked_at":       lastChecked,
				"consecutive_failures":  prx.ConsecutiveFailures,
				"total_checks":          prx.TotalChecks,
				"uptime_percentage":     fmt.Sprintf("%.1f", uptimePct),
				"history":              histEntries,
			}
			json.NewEncoder(w).Encode(resp)
		}
	}
}
