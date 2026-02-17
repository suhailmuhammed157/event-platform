package kafka

import (
	"context"
	"errors"

	"event-platform/pkg/event"
	"event-platform/pkg/observability"

	"time"

	"github.com/segmentio/kafka-go"
)

type BatchConsumer struct {
	reader       *kafka.Reader
	pool         *Pool
	batchSize    int
	batchTimeout time.Duration
	producer     *MultiTopicProducer
	handler      JobHandler
}

func NewBatchConsumer(broker, topic, groupID string, pool *Pool, batchSize int, batchTimeout time.Duration, handler JobHandler, producer *MultiTopicProducer) (*BatchConsumer, error) {

	if producer == nil {
		return nil, errors.New("all producers must be non-nil")
	}
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
		producer:     producer,
	}, nil
}

func (c *BatchConsumer) Run(ctx context.Context) error {
	batch := make([]kafka.Message, 0, c.batchSize)
	ticker := time.NewTicker(c.batchTimeout)
	defer ticker.Stop()

	msgCh := make(chan kafka.Message, 100)

	go func() {
		defer close(msgCh)
		for {
			m, err := c.reader.FetchMessage(ctx)
			if err != nil {
				return
			}
			msgCh <- m
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			if len(batch) > 0 {
				c.flush(ctx, batch)
				batch = batch[:0]
			}

		case m := <-msgCh:
			batch = append(batch, m)

			if len(batch) >= c.batchSize {
				c.flush(ctx, batch)
				batch = batch[:0]
			}
		}
	}
}

func (c *BatchConsumer) flush(parentCtx context.Context, batch []kafka.Message) {

	for _, msg := range batch {
		if msg.Value == nil {
			continue
		}

		m := msg

		c.pool.Submit(func(_ context.Context) error {
			start := time.Now()
			err := c.handler.Handle(parentCtx, string(m.Key), m.Value)
			observability.ProcessingLatency.Observe(time.Since(start).Seconds())

			// select topic
			topic := "events.retry"
			if err == nil {
				topic = "events.processed"
			} else if errors.Is(err, event.ErrFatal) {
				topic = "events.dlq"
			}

			if err := c.producer.Publish(parentCtx, topic, string(m.Key), m.Value); err != nil {
				return err
			}

			if err := c.reader.CommitMessages(parentCtx, m); err == nil {
				switch topic {
				case "events.processed":
					observability.ProcessedEvents.Inc()
				case "events.retry":
					observability.RetryEvents.Inc()
				case "events.dlq":
					observability.DLQEvents.Inc()
				}
			}

			return err
		})
	}

}

func (c *BatchConsumer) Close() error {
	return c.reader.Close()
}
