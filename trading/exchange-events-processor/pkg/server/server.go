package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"exchange-events-processor/pkg/config"
	kafka_consumer "exchange-events-processor/pkg/kafka/consumer"
)

// Start initializes dependencies, sets up HTTP and Kafka handlers, and starts the server.
func Start(ctx context.Context, logger *log.Logger, shutdownChannel chan struct{}) {
	logger.Println("Server is starting...")

	// Load dependencies
	err := loadDependencies(ctx, logger)
	defer closeDependencies(ctx, logger)
	if err != nil {
		logger.Fatalf("Loading dependencies failed: %v", err)
		return
	}

	logger.Printf("Kafka topics: Order=%s, Trades=%s, CashBalance=%s",
		config.KafkaConfig.OrderTopic, config.KafkaConfig.TradesTopic, config.KafkaConfig.CashBalanceTopic)

	// Register Kafka message handlers
	kafka_consumer.GlobalConsumer.SetMessageHandler(config.KafkaConfig.TradesTopic, func(msg kafka_consumer.Message) error {
		return TradeHandler(ctx, msg)
	})
	kafka_consumer.GlobalConsumer.SetMessageHandler(config.KafkaConfig.CashBalanceTopic, func(msg kafka_consumer.Message) error {
		return CashBalanceHandler(ctx, msg)
	})
	kafka_consumer.GlobalConsumer.SetMessageHandler(config.KafkaConfig.OrderTopic, func(msg kafka_consumer.Message) error {
		return OrderHandler(ctx, msg)
	})

	// Start Kafka consumer in a separate goroutine
	go kafka_consumer.GlobalConsumer.Pull(ctx)

	// Setup HTTP server and handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/health", HealthHandler)
	mux.HandleFunc("/health/kafka", KafkaHealthHandler)

	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", config.ServerConfig.Port),
		Handler: mux,
	}

	// Start HTTP server in a separate goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("Server crashed: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-shutdownChannel
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Failed to shutdown server gracefully: %v", err)
		return
	}

	logger.Println("Exiting after gracefully closing dependencies")
}
