package kafka

import (
	"context"
	"errors"

	"event-platform/pkg/event"

	"time"

	"github.com/segmentio/kafka-go"
)

type BatchConsumer struct {
	reader        *kafka.Reader
	pool          *Pool
	batchSize     int
	batchTimeout  time.Duration
	retryProducer *MultiTopicProducer
	dlqProducer   *MultiTopicProducer
	handler       JobHandler
}

func NewBatchConsumer(broker, topic, groupID string, pool *Pool, batchSize int, batchTimeout time.Duration, handler JobHandler) *BatchConsumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{broker},
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	return &BatchConsumer{
		reader:       r,
		pool:         pool,
		batchSize:    batchSize,
		batchTimeout: batchTimeout,
		handler:      handler,
	}
}

func (c *BatchConsumer) Run(ctx context.Context) error {
	batch := make([]kafka.Message, 0, c.batchSize)
	ticker := time.NewTicker(c.batchTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if len(batch) > 0 {
				c.flush(ctx, batch)
				batch = batch[:0]
			}
		default:
			m, err := c.reader.FetchMessage(ctx)
			if err != nil {
				return err
			}
			batch = append(batch, m)

			if len(batch) >= c.batchSize {
				c.flush(ctx, batch)
				batch = batch[:0]
			}
		}
	}
}

func (c *BatchConsumer) flush(ctx context.Context, batch []kafka.Message) {
	for _, m := range batch {
		msg := m
		c.pool.Submit(func(ctx context.Context) error {
			err := c.handler.Handle(ctx, string(msg.Key), msg.Value)
			if err == nil {
				return c.reader.CommitMessages(ctx, msg)
			}

			if errors.Is(err, event.ErrRetryable) {
				_ = c.retryProducer.Publish(ctx, "events.retry", string(msg.Key), msg.Value)
				return c.reader.CommitMessages(ctx, msg)
			}

			_ = c.dlqProducer.Publish(ctx, "events.dlq", string(msg.Key), msg.Value)
			return c.reader.CommitMessages(ctx, msg)
		})
	}
}

func (c *BatchConsumer) Close() error {
	return c.reader.Close()
}
