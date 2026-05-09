package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/HasithaErandika/proxy-maze/internal/integration"
)

// IntegrationsHandler handles POST /integrations
func IntegrationsHandler(registry *integration.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// First validate JSON structure
		var raw json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			http.Error(w, `{"error":"malformed JSON"}`, http.StatusBadRequest)
			return
		}

		var req integration.IntegrationReq
		json.Unmarshal(raw, &req)

		if req.Type == "" || req.WebhookURL == "" {
			http.Error(w, "Missing type or webhook_url", http.StatusBadRequest)
			return
		}

		ig := integration.Integration{
			Type:       req.Type,
			WebhookURL: req.WebhookURL,
			Username:   req.Username,
			Events:     req.Events,
		}

		registry.Add(ig)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(ig)
	}
}
