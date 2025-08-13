package server

import (
	"context"
	"log"

	"exchange-events-processor/pkg/kafka"
	"exchange-events-processor/pkg/kafka/consumer"
)

var (
	loadKafka = kafka.LoadKafka
)

// loadDependencies initializes all required dependencies.
func loadDependencies(ctx context.Context, logger *log.Logger) error {
	if err := loadKafka(ctx, logger); err != nil {
		return err
	}
	return nil
}

// closeDependencies closes all initialized dependencies gracefully.
func closeDependencies(ctx context.Context, logger *log.Logger) {
	if consumer.Consumer != nil {
		if err := consumer.Consumer.KafkaConsumer.Close(); err != nil {
			logger.Fatalf("Failure closing Kafka consumer gracefully: %v", err)
		}
	}
}

