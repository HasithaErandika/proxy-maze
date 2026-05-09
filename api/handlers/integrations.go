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

		var req integration.IntegrationReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// Ignore parse errors for unknown fields
		}

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

		w.WriteHeader(http.StatusCreated)
	}
}
