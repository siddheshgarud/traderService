package server

import (
	"context"
	"log"

	"exchange-events-producer/pkg/kafka"
)

// loadKafka is an alias for kafka.LoadKafka for easier testing/mocking.
var (
	loadKafka = kafka.LoadKafka
)

// loadDependencies loads all required dependencies for the server.
func loadDependencies(ctx context.Context, logger *log.Logger) error {
	// Load Kafka
	if err := loadKafka(ctx, logger); err != nil {
		return err
	}

	return nil
}
