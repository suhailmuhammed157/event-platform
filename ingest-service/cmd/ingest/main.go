package main

import (
	"event-platform/ingest-service/internal/batching"
	"event-platform/ingest-service/internal/config"
	"event-platform/ingest-service/internal/http"
	"event-platform/ingest-service/internal/producer"
	"time"
)

func main() {
	cfg := config.LoadConfig()
	kafkaProducer := producer.NewKafkaProducer(cfg.KafkaBroker, cfg.KafkaTopic)
	defer kafkaProducer.Close()

	batcher := batching.NewBatcher(kafkaProducer, 50, 10*time.Millisecond)
	defer batcher.Close()

	server := http.NewServer(batcher)
	server.Run(cfg.Port)
}
