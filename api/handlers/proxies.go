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
			var req ProxiesReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				// Accept unknown fields silently, but check for valid JSON format
			}

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

			for _, p := range all {
				if p.Status == proxy.StatusUp {
					up++
				} else if p.Status == proxy.StatusDown {
					down++
				}
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
				"proxies":      all,
			})

		case http.MethodDelete:
			pool.Clear()
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	}
}
