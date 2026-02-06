package batching

import (
	"context"
	"event-platform/pkg/kafka"
	"log"
	"sync"
	"time"
)

type Event struct {
	Key   string
	Value []byte
	Topic string
}

type Batcher struct {
	producer     *kafka.MultiTopicProducer
	batch        []Event
	batchSize    int
	batchTimeout time.Duration
	mutex        sync.Mutex
	ticker       *time.Ticker
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewBatcher(p *kafka.MultiTopicProducer, batchSize int, batchTimeout time.Duration) *Batcher {
	ctx, cancel := context.WithCancel(context.Background())
	b := &Batcher{
		producer:     p,
		batch:        make([]Event, 0, batchSize),
		batchSize:    batchSize,
		batchTimeout: batchTimeout,
		ticker:       time.NewTicker(batchTimeout),
		ctx:          ctx,
		cancel:       cancel,
	}

	go b.run()
	return b
}

// Add event to batch
func (b *Batcher) Add(event Event) {
	b.mutex.Lock()
	b.batch = append(b.batch, event)
	if len(b.batch) >= b.batchSize {
		b.flushLocked()
	}
	b.mutex.Unlock()
}

// Run flush on timeout
func (b *Batcher) run() {
	for {
		select {
		case <-b.ctx.Done():
			return
		case <-b.ticker.C:
			b.mutex.Lock()
			b.flushLocked()
			b.mutex.Unlock()
		}
	}
}

// Flush batch to Kafka
func (b *Batcher) flushLocked() {
	if len(b.batch) == 0 {
		return
	}

	batchCopy := b.batch
	b.batch = make([]Event, 0, b.batchSize)

	go func(events []Event) {
		for _, e := range events {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			if err := b.producer.Publish(ctx, e.Topic, e.Key, e.Value); err != nil {
				log.Printf("failed to publish event %s: %v", e.Key, err)
			}
			cancel()
		}
	}(batchCopy)
}

func (b *Batcher) Close() {
	b.cancel()
	b.ticker.Stop()
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.flushLocked()
}
