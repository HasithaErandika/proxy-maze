package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/HasithaErandika/proxy-maze/internal/webhook"
)

type WebhookReq struct {
	URL string `json:"url"`
}

func WebhooksHandler(registry *webhook.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var req WebhookReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		}

		if req.URL == "" {
			http.Error(w, "Missing URL", http.StatusBadRequest)
			return
		}

		wh := registry.Add(req.URL)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"webhook_id": wh.ID,
			"url":        wh.URL,
		})
	}
}
