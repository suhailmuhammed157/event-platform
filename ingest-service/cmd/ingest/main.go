package main

import (
	"event-platform/ingest-service/internal/batching"
	"event-platform/ingest-service/internal/config"
	"event-platform/ingest-service/internal/http"
	"event-platform/pkg/kafka"
	"event-platform/pkg/observability"

	"time"
)

func main() {
	cfg := config.LoadConfig()

	// 1️⃣ Metrics
	observability.Init()
	observability.ServeMetrics("8081") // metrics exposed at http://localhost:8081/metrics

	kafkaMultiTopicProducer := kafka.NewMultiTopicProducer(cfg.KafkaBroker, cfg.KafkaTopic)
	defer kafkaMultiTopicProducer.Close()

	batcher := batching.NewBatcher(kafkaMultiTopicProducer, 50, 10*time.Millisecond)
	defer batcher.Close()

	server := http.NewServer(batcher)
	server.Run(cfg.Port)
}
