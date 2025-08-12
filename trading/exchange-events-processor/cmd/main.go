package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"exchange-events-processor/pkg/config"
	"exchange-events-processor/pkg/server"
)

func main() {

	shutdownChannel := make(chan struct{})

	signals := make(chan os.Signal, 1)
	signal.Notify(signals,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := log.Default()

	if err := config.Load(); err != nil {
		logger.Fatalf("error initializing application config, Error: %s", err.Error())
	}

	go func() {
		<-signals
		logger.Println("Initialize graceful shutdown of web server")
		close(shutdownChannel)
	}()

	server.Start(ctx, logger, shutdownChannel)

}
