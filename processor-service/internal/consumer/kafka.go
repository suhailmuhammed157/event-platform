package consumer

import (
	"context"

	"github.com/segmentio/kafka-go"

	"event-platform/processor-service/internal/handler"
	"event-platform/processor-service/internal/workers"
)

type Consumer struct {
	reader *kafka.Reader
	pool   *workers.Pool
}

func NewConsumer(broker, topic, groupID string, pool *workers.Pool) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{broker},
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &Consumer{
		reader: r,
		pool:   pool,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			return err
		}

		// Submit job to worker pool
		key := string(m.Key)
		value := m.Value

		c.pool.Submit(func(ctx context.Context) error {
			if err := handler.HandleEvent(ctx, key, value); err != nil {
				return err
			}
			return c.reader.CommitMessages(ctx, m)
		})
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
