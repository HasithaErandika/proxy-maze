package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/HasithaErandika/proxy-maze/internal/proxy"
)

type ProxiesReq struct {
	Proxies []string `json:"proxies"`
	Replace bool     `json:"replace"`
}

// ProxiesHandler handles POST, GET, DELETE for /proxies
func ProxiesHandler(pool *proxy.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var raw json.RawMessage
			if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
				http.Error(w, `{"error":"malformed JSON"}`, http.StatusBadRequest)
				return
			}

			var req ProxiesReq
			json.Unmarshal(raw, &req)

			added := pool.Add(req.Proxies, req.Replace)

			// The newly added proxies are returned as [{id, url, status:"pending"}]
			respProxies := make([]map[string]string, 0, len(added))
			for _, p := range added {
				respProxies = append(respProxies, map[string]string{
					"id":     p.ID,
					"url":    p.URL,
					"status": p.Status,
				})
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"accepted": len(added),
				"proxies":  respProxies,
			})

		case http.MethodGet:
			all := pool.GetAll()
			total := len(all)
			up := 0
			down := 0

			type proxyView struct {
				ID                  string  `json:"id"`
				URL                 string  `json:"url"`
				Status              string  `json:"status"`
				LastCheckedAt       *string `json:"last_checked_at"`
				ConsecutiveFailures int     `json:"consecutive_failures"`
			}

			proxyViews := make([]proxyView, 0, total)
			for _, p := range all {
				if p.Status == proxy.StatusUp {
					up++
				} else if p.Status == proxy.StatusDown {
					down++
				}
				var lastChecked *string
				if p.LastCheckedAt != nil {
					s := p.LastCheckedAt.UTC().Format("2006-01-02T15:04:05Z")
					lastChecked = &s
				}
				proxyViews = append(proxyViews, proxyView{
					ID:                  p.ID,
					URL:                 p.URL,
					Status:              p.Status,
					LastCheckedAt:       lastChecked,
					ConsecutiveFailures: p.ConsecutiveFailures,
				})
			}

			failureRate := 0.0
			if total > 0 {
				failureRate = float64(down) / float64(total)
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"total":        total,
				"up":           up,
				"down":         down,
				"failure_rate": failureRate,
				"proxies":      proxyViews,
			})

		case http.MethodDelete:
			pool.Clear()
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	}
}
