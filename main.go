package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
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

	alertManager := alert.NewManager(dispatcher)
	alertManager.SetIntegrationRegistry(igRegistry)
	
	// Complete metrics setup
	metricsTracker = metrics.NewMetrics(pool, alertManager)
	dispatcher.SetMetrics(metricsTracker)

	checker := proxy.NewChecker(pool, cfgStore, alertManager, metricsTracker)

	// Mutex to protect cancel and context variables
	var mu sync.Mutex
	ctx, cancel := context.WithCancel(context.Background())

	// Link config store to checker loop restart
	restartLoop := func() {
		mu.Lock()
		cancel()
		ctx, cancel = context.WithCancel(context.Background())
		go checker.Start(ctx)
		mu.Unlock()
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
	mu.Lock()
	cancel()
	mu.Unlock()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Shutdown error: %v", err)
	}
	log.Println("Shutdown complete.")
}
