package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"exchange-events-producer/pkg/config"
	"exchange-events-producer/pkg/kafka/producer"
)

// Start initializes and runs the HTTP server, handling graceful shutdown.
func Start(ctx context.Context, logger *log.Logger, shutdownChannel chan struct{}) {
	logger.Println("Server is starting...")

	// Load dependencies
	err := loadDependencies(ctx, logger)
	if err != nil {
		logger.Fatalf("Loading dependencies failed: %v", err)
		return
	}

	// Set up HTTP handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/health", HealthHandler)
	mux.HandleFunc("/produce-random", ProduceRandomHandler)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.ServerConfig.Port),
		Handler: mux,
	}

	// Ensure producer is closed on exit
	defer func() {
		producer.GlobalProducer.Close()
		logger.Println("Exited after gracefully closing dependencies")
	}()

	// Start server in a goroutine
	go func() {
		logger.Printf("Listening on port %d", config.ServerConfig.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("Server crashed: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-shutdownChannel
	logger.Println("Shutdown signal received, shutting down server...")

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Failed to shutdown server gracefully: %v", err)
	}
}
