package config

import (
	"os"
)

type Config struct {
	KafkaBroker string
	KafkaTopic  string
	Port        string
}

func LoadConfig() Config {
	return Config{
		KafkaBroker: getEnv("KAFKA_BROKER", "localhost:9092"),
		KafkaTopic:  getEnv("KAFKA_TOPIC", "events"),
		Port:        getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return fallback
}
