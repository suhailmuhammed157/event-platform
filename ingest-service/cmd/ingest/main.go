package main

import (
	"event-platform/ingest-service/internal/batching"
	"event-platform/ingest-service/internal/config"
	"event-platform/ingest-service/internal/http"
	"event-platform/pkg/kafka"

	"time"
)

func main() {
	cfg := config.LoadConfig()
	kafkaMultiTopicProducer := kafka.NewMultiTopicProducer(cfg.KafkaBroker, cfg.KafkaTopic)
	defer kafkaMultiTopicProducer.Close()

	batcher := batching.NewBatcher(kafkaMultiTopicProducer, 50, 10*time.Millisecond)
	defer batcher.Close()

	server := http.NewServer(batcher)
	server.Run(cfg.Port)
}
