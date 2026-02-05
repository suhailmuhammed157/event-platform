package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"event-platform/processor-service/internal/config"
	"event-platform/processor-service/internal/consumer"
	"event-platform/processor-service/internal/workers"
)

func main() {
	cfg := config.LoadConfig()

	pool := workers.NewPool(cfg.WorkerCount, 1024)
	defer pool.Shutdown()

	c := consumer.NewConsumer(cfg.KafkaBroker, cfg.KafkaTopic, cfg.GroupID, pool)
	defer c.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("processor service started")

	if err := c.Run(ctx); err != nil {
		log.Printf("consumer stopped: %v", err)
	}
}
