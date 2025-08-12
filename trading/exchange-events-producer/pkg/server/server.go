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




func Start(ctx context.Context, logger *log.Logger, shutdownChannel chan struct{}) {
	logger.Println("Server is starting...")

	err := loadDependencies(ctx, logger)
	if err != nil {
		log.Fatalf("Loading dependencies failed", err.Error())
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", HealthHandler)
	mux.HandleFunc("/produce-random", ProduceRandomHandler)

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
	producer.GlobalProducer.Close()
	log.Printf("existing after gracefully closing dependencies")
}
