package handlers

import (
	"encoding/json"
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
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if isHistory {
			json.NewEncoder(w).Encode(prx.History)
		} else {
			json.NewEncoder(w).Encode(prx)
		}
	}
}
