package main

import (
	"event-platform/ingest-service/internal/config"
	"event-platform/ingest-service/internal/http"
	"event-platform/ingest-service/internal/producer"
)

func main() {
	cfg := config.LoadConfig()
	kafkaProducer := producer.NewKafkaProducer(cfg.KafkaBroker, cfg.KafkaTopic)
	defer kafkaProducer.Close()

	server := http.NewServer(kafkaProducer)
	server.Run(cfg.Port)
}
