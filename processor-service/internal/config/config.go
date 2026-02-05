package config

import (
	"os"
	"strconv"
)

type Config struct {
	KafkaBroker string
	KafkaTopic  string
	GroupID     string
	WorkerCount int
}

func LoadConfig() Config {
	return Config{
		KafkaBroker: getEnv("KAFKA_BROKER", "localhost:9092"),
		KafkaTopic:  getEnv("KAFKA_TOPIC", "events"),
		GroupID:     getEnv("KAFKA_GROUP_ID", "event-processor"),
		WorkerCount: getEnvInt("WORKERS", 8),
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
