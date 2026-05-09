package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/HasithaErandika/proxy-maze/internal/config"
)

type ConfigReq struct {
	CheckIntervalSeconds int `json:"check_interval_seconds"`
	RequestTimeoutMs     int `json:"request_timeout_ms"`
}

// ConfigHandler handles POST and GET for /config
func ConfigHandler(cfgStore *config.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			interval, timeout := cfgStore.Get()
			resp := ConfigReq{
				CheckIntervalSeconds: interval,
				RequestTimeoutMs:     timeout,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		if r.Method == http.MethodPost {
			var raw json.RawMessage
			if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
				http.Error(w, `{"error":"malformed JSON"}`, http.StatusBadRequest)
				return
			}

			var req ConfigReq
			// Unmarshal from raw — this will silently ignore unknown fields
			json.Unmarshal(raw, &req)

			// Validate or use defaults
			if req.CheckIntervalSeconds <= 0 {
				req.CheckIntervalSeconds = 30
			}
			if req.RequestTimeoutMs <= 0 {
				req.RequestTimeoutMs = 5000
			}

			cfgStore.Update(req.CheckIntervalSeconds, req.RequestTimeoutMs)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(req)
			return
		}

		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
