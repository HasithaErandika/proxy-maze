package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/HasithaErandika/proxy-maze/internal/integration"
)

func IntegrationsHandler(registry *integration.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var req integration.IntegrationReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		}

		if req.Type == "" || req.WebhookURL == "" {
			http.Error(w, "Missing type or webhook_url", http.StatusBadRequest)
			return
		}

		username := req.Username
		if username == "" {
			username = "ProxyWatch"
		}

		events := req.Events
		if len(events) == 0 {
			events = []string{"alert.fired", "alert.resolved"}
		}

		ig := integration.Integration{
			Type:       req.Type,
			WebhookURL: req.WebhookURL,
			Username:   username,
			Events:     events,
		}

		registry.Add(ig)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(ig)
	}
}
