package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"event-platform/pkg/observability"
	"event-platform/sink-service/internal/config"
	"event-platform/sink-service/internal/storage"

	"github.com/segmentio/kafka-go"
)

func main() {
	cfg := config.Load()

	// 1️⃣ Metrics
	observability.Init()
	observability.ServeMetrics("8083") // metrics exposed at http://localhost:8083/metrics

	sink, err := storage.NewPostgresSink(cfg.PostgresDSN)
	if err != nil {
		log.Fatal(err)
	}

	err = storage.RunMigrations(cfg.PostgresDSN, "./migrations")
	if err != nil {
		log.Fatal(err)
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{cfg.KafkaBroker},
		Topic:   cfg.KafkaTopic,
		GroupID: cfg.GroupID,
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("sink service started")

	batch := make([]storage.Event, 0, cfg.BatchSize)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	msgCh := make(chan kafka.Message, 500)

	go func() {
		defer close(msgCh)
		for {
			m, err := reader.FetchMessage(ctx)
			if err != nil {
				return
			}
			msgCh <- m
		}
	}()

	for {

		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			if len(batch) > 0 {
				flush(ctx, sink, reader, batch)
				batch = batch[:0]
			}
		case m := <-msgCh:
			batch = append(batch, storage.Event{
				ID:      string(m.Key),
				Payload: m.Value,
				Msg:     m,
			})

			if len(batch) >= cfg.BatchSize {
				flush(ctx, sink, reader, batch)
				batch = batch[:0]

			}
		}
	}
}

func flush(ctx context.Context, sink *storage.PostgresSink, reader *kafka.Reader, batch []storage.Event) {

	if err := sink.InsertBatch(ctx, batch); err != nil {
		log.Printf("batch insert failed: %v", err)
		return
	}

	// once the batch insert is success, then commit all the kafka message
	msgs := make([]kafka.Message, 0, len(batch))
	for _, e := range batch {
		msgs = append(msgs, e.Msg)
	}

	if err := reader.CommitMessages(ctx, msgs...); err != nil {
		log.Printf("kafka commit failed: %v", err)
	}
}
