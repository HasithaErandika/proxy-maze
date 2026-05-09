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
	cfgStore := config.NewStore()
	pool := proxy.NewPool()
	whRegistry := webhook.NewRegistry()
	igRegistry := integration.NewRegistry()
	
	metricsTracker := metrics.NewMetrics(pool, nil)

	dispatcher := webhook.NewDispatcher(whRegistry, metricsTracker)

	alertManager := alert.NewManager(dispatcher)
	alertManager.SetIntegrationRegistry(igRegistry) 
	
	metricsTracker = metrics.NewMetrics(pool, alertManager)
	dispatcher.SetMetrics(metricsTracker)

	checker := proxy.NewChecker(pool, cfgStore, alertManager, metricsTracker)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	restartLoop := func() {
		cancel()
		ctx, cancel = context.WithCancel(context.Background())
		go checker.Start(ctx)
	}
	cfgStore.SetTickerCancel(restartLoop)

	go checker.Start(ctx)

	mux := api.NewRouter(cfgStore, pool, alertManager, whRegistry, igRegistry, metricsTracker)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

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
