package config

import "os"

type Config struct {
	KafkaBroker string
	KafkaTopic  string
	GroupID     string
	PostgresDSN string
	BatchSize   int
}

func Load() Config {
	return Config{
		KafkaBroker: getenv("KAFKA_BROKER", "localhost:9092"),
		KafkaTopic:  getenv("KAFKA_TOPIC", "events.processed"),
		GroupID:     getenv("KAFKA_GROUP", "sink-service"),
		PostgresDSN: getenv("POSTGRES_DSN", "postgres://user:pass@localhost:5532/events_db?sslmode=disable"),
		BatchSize:   100,
	}
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
