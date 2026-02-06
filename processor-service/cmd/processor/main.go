package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"event-platform/pkg/kafka"
	"event-platform/processor-service/internal/config"
	"event-platform/processor-service/internal/handler"
)

func main() {
	cfg := config.LoadConfig()

	pool := kafka.NewPool(cfg.WorkerCount, 1024)
	defer pool.Shutdown()

	batchConsumer := kafka.NewBatchConsumer(
		cfg.KafkaBroker,
		cfg.KafkaTopic,
		cfg.GroupID,
		pool,
		100,                                    // batch size
		20*time.Millisecond,                    // batch timeout
		kafka.HandlerFunc(handler.HandleEvent), // <-- wrap function here
	)
	defer batchConsumer.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("processor service started")

	if err := batchConsumer.Run(ctx); err != nil {
		log.Printf("consumer stopped: %v", err)
	}
}
