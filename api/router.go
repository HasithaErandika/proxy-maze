package api

import (
	"net/http"

	"github.com/HasithaErandika/proxy-maze/api/handlers"
	"github.com/HasithaErandika/proxy-maze/internal/alert"
	"github.com/HasithaErandika/proxy-maze/internal/config"
	"github.com/HasithaErandika/proxy-maze/internal/integration"
	"github.com/HasithaErandika/proxy-maze/internal/metrics"
	"github.com/HasithaErandika/proxy-maze/internal/proxy"
	"github.com/HasithaErandika/proxy-maze/internal/webhook"
)

func NewRouter(
	cfgStore *config.Store,
	pool *proxy.Pool,
	alertManager *alert.Manager,
	whRegistry *webhook.Registry,
	igRegistry *integration.Registry,
	metricsTracker *metrics.Metrics,
) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", handlers.HealthHandler)
	mux.HandleFunc("/config", handlers.ConfigHandler(cfgStore))
	mux.HandleFunc("/proxies", handlers.ProxiesHandler(pool))
	mux.HandleFunc("/proxies/", handlers.ProxyHandler(pool))
	mux.HandleFunc("/alerts", handlers.AlertsHandler(alertManager))
	mux.HandleFunc("/webhooks", handlers.WebhooksHandler(whRegistry))
	mux.HandleFunc("/integrations", handlers.IntegrationsHandler(igRegistry))
	mux.HandleFunc("/metrics", handlers.MetricsHandler(metricsTracker))

	return mux
}
