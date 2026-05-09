package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/HasithaErandika/proxy-maze/internal/alert"
)

func AlertsHandler(manager *alert.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		alerts := manager.GetAll()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(alerts)
	}
}
