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

type ()

func Start(ctx context.Context, logger *log.Logger, shutdownChannel chan struct{}) {
	logger.Println("Server is starting...")

	err := loadDependencies(ctx, logger)
	defer closeDependencies(ctx, logger)
	if err != nil {
		log.Fatalf("Loading dependencies failed", err.Error())
		return
	}
	fmt.Println("order topic - ", config.KafkaConfig.OrderTopic)
kafka_consumer.GlobalConsumer.SetMessageHandler(config.KafkaConfig.OrderTopic, func(msg kafka_consumer.Message) error {
        return OrderHandler(ctx, msg)
    })
    kafka_consumer.GlobalConsumer.SetMessageHandler(config.KafkaConfig.TradesTopic, func(msg kafka_consumer.Message) error {
        return TradeHandler(ctx, msg)
    })
    kafka_consumer.GlobalConsumer.SetMessageHandler(config.KafkaConfig.CashBalanceTopic, func(msg kafka_consumer.Message) error {
        return CashBalanceHandler(ctx, msg)
    })
	
	go kafka_consumer.GlobalConsumer.Pull(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", HealthHandler)
	mux.HandleFunc("/health/kafka", KafkaHealthHandler)

	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", config.ServerConfig.Port),
		Handler: mux,
	}

	go func() {
		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Server Crashed")
			return
		}
	}()

	<-shutdownChannel
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("failed to shutdown server gracefully")
		return

	}

	log.Printf("existing after gracefully closing dependencies")
}
