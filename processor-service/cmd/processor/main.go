package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"event-platform/pkg/kafka"
	"event-platform/pkg/observability"
	"event-platform/pkg/pprof"
	"event-platform/processor-service/internal/config"
	"event-platform/processor-service/internal/handler"
)

func main() {
	cfg := config.LoadConfig()

	pprof.Start("6060")

	// 1️⃣ Metrics
	observability.Init()
	observability.ServeMetrics("8082") // metrics exposed at http://localhost:8082/metrics

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool := kafka.NewPool(ctx, cfg.WorkerCount, 10000)
	defer func() {
		log.Println("shutting down pool...")
		pool.Shutdown()
	}()

	producer := kafka.NewMultiTopicProducer(cfg.KafkaBroker, "events.processed",
		"events.retry",
		"events.dlq")

	batchConsumer, err := kafka.NewBatchConsumer(
		cfg.KafkaBroker,
		cfg.KafkaTopic,
		cfg.GroupID,
		pool,
		1000,               // batch size
		1*time.Millisecond, // batch timeout
		kafka.HandlerFunc(handler.HandleEvent),
		producer,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer batchConsumer.Close()

	log.Println("processor service started")

	if err := batchConsumer.Run(ctx); err != nil {
		log.Printf("consumer stopped: %v", err)
	}
}
