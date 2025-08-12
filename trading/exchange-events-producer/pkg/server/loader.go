package server

import (
	"context"
	"log"

	"exchange-events-producer/pkg/kafka"

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
