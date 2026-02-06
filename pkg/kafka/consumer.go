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

	const backPressureThreshold = 0.8 // 80% of worker queue

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
			// 🔥 BackPressure: slow down consumption if pool queue is almost full
			queueLen := c.pool.QueueLen()
			queueCap := c.pool.Capacity()

			if float64(queueLen)/float64(queueCap) > backPressureThreshold {
				time.Sleep(5 * time.Millisecond)
				continue
			}

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

func (c *BatchConsumer) flush(parentCtx context.Context, batch []kafka.Message) {
	for _, m := range batch {
		msg := m

		// derive per-job context (can add timeout here if needed)
		jobCtx, cancel := context.WithCancel(parentCtx)

		c.pool.Submit(func(_ context.Context) error {
			defer cancel()

			err := c.handler.Handle(jobCtx, string(msg.Key), msg.Value)
			if err == nil {
				return c.reader.CommitMessages(jobCtx, msg)
			}

			if errors.Is(err, event.ErrRetryable) {
				_ = c.retryProducer.Publish(jobCtx, "events.retry", string(msg.Key), msg.Value)
				return c.reader.CommitMessages(jobCtx, msg)
			}

			_ = c.dlqProducer.Publish(jobCtx, "events.dlq", string(msg.Key), msg.Value)
			return c.reader.CommitMessages(jobCtx, msg)
		})
	}
}

func (c *BatchConsumer) Close() error {
	return c.reader.Close()
}
