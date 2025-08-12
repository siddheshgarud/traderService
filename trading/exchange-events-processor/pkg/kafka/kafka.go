package kafka

import (
	"context"
	"exchange-events-processor/pkg/kafka/consumer"
	"log"
)


func LoadKafka(ctx context.Context, log *log.Logger) error {
	if err := consumer.LoadConsumer(ctx, *log); err != nil {
		return err
	}

	return nil
}