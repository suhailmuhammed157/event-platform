package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

func main() {
	// CLI flags
	broker := flag.String("broker", "localhost:9092", "Kafka broker address")
	topic := flag.String("topic", "events", "Kafka topic to publish to")
	rps := flag.Int("rps", 1000, "Requests per second")
	duration := flag.Int("duration", 10, "Duration in seconds")
	flag.Parse()

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{*broker},
		Topic:    *topic,
		Balancer: &kafka.LeastBytes{},
	})
	defer writer.Close()

	fmt.Printf("Starting loadgen: %d RPS for %d seconds to topic %s\n", *rps, *duration, *topic)

	ctx := context.Background()
	ticker := time.NewTicker(time.Second / time.Duration(*rps))
	defer ticker.Stop()

	end := time.Now().Add(time.Duration(*duration) * time.Second)
	count := 0

	for time.Now().Before(end) {
		<-ticker.C
		count++

		go func() {
			key := uuid.New().String()
			value := []byte(fmt.Sprintf(`{"id":"%s","ts":%d}`, key, time.Now().UnixMilli()))
			msg := kafka.Message{
				Key:   []byte(key),
				Value: value,
			}

			if err := writer.WriteMessages(ctx, msg); err != nil {
				log.Printf("failed to write message: %v", err)
			}
		}()
	}

	fmt.Printf("Load generation complete: %d messages sent\n", count)
}
