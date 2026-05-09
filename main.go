package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HasithaErandika/proxy-maze/api"
	"github.com/HasithaErandika/proxy-maze/internal/alert"
	"github.com/HasithaErandika/proxy-maze/internal/config"
	"github.com/HasithaErandika/proxy-maze/internal/integration"
	"github.com/HasithaErandika/proxy-maze/internal/metrics"
	"github.com/HasithaErandika/proxy-maze/internal/proxy"
	"github.com/HasithaErandika/proxy-maze/internal/webhook"
)

func main() {
	// Initialize core components
	cfgStore := config.NewStore()
	pool := proxy.NewPool()
	whRegistry := webhook.NewRegistry()
	igRegistry := integration.NewRegistry()
	
	// Pre-create metrics struct
	metricsTracker := metrics.NewMetrics(pool, nil)

	// Webhook dispatcher needs metrics tracking
	dispatcher := webhook.NewDispatcher(whRegistry, metricsTracker)

	// Add integrations to dispatcher via wrapping to avoid circular dependencies
	// Actually, wait, let's inject integration registry into alert manager or just use a wrapper.
	// For simplicity, we just inject dispatcher into alert manager, and we will update manager to also trigger integrations.
	alertManager := alert.NewManager(dispatcher)
	alertManager.SetIntegrationRegistry(igRegistry) // We'll add this method to alertManager
	
	// Complete metrics setup
	metricsTracker = metrics.NewMetrics(pool, alertManager)
	dispatcher.SetMetrics(metricsTracker) // update dispatcher with real metrics

	checker := proxy.NewChecker(pool, cfgStore, alertManager, metricsTracker)

	// Context for graceful shutdown and background loop cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Link config store to checker loop
	restartLoop := func() {
		cancel()
		ctx, cancel = context.WithCancel(context.Background())
		go checker.Start(ctx)
	}
	cfgStore.SetTickerCancel(restartLoop)

	// Start background loop for the first time
	go checker.Start(ctx)

	// API Routing
	mux := api.NewRouter(cfgStore, pool, alertManager, whRegistry, igRegistry, metricsTracker)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Graceful shutdown handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Starting server on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Shutdown error: %v", err)
	}
	log.Println("Shutdown complete.")
}
