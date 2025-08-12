package kafka

import (
	"context"
	"exchange-events-producer/pkg/kafka/producer"
	"log"
)


func LoadKafka(ctx context.Context, log *log.Logger) error {
	if err := producer.LoadProducer(ctx, log); err != nil {
		return err
	}

	return nil
}