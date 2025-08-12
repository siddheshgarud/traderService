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

func loadDependencies(ctx context.Context, log *log.Logger) error {

	// Load Kafka
	if err := loadKafka(ctx, log); err != nil {
		return err
	}

	return nil
}

func closeDependencies(ctx context.Context, log *log.Logger) {
	if consumer.Consumer != nil {
		if err := consumer.Consumer.KafkaConsumer.Close(); err != nil {
			log.Fatalf("Failure Closing Kafka Consumer Gracefully")
		}
	}
}
