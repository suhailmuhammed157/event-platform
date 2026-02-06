package kafka

import (
	"context"
	"errors"
	"log"

	"github.com/segmentio/kafka-go"
)

var ErrUnknownTopic = errors.New("unknown kafka topic")

type MultiTopicProducer struct {
	writers map[string]*kafka.Writer
}

func NewMultiTopicProducer(broker string, topics ...string) *MultiTopicProducer {
	writers := make(map[string]*kafka.Writer)
	for _, topic := range topics {
		writers[topic] = &kafka.Writer{
			Addr:  kafka.TCP(broker),
			Topic: topic,
		}
	}
	return &MultiTopicProducer{writers: writers}
}

func (p *MultiTopicProducer) Publish(ctx context.Context, topic, key string, value []byte) error {
	w, ok := p.writers[topic]
	if !ok {
		return ErrUnknownTopic
	}
	return w.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: value,
	})
}

func (p *MultiTopicProducer) Close() error {
	var firstErr error
	for topic, w := range p.writers {
		if err := w.Close(); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			log.Printf("failed to close kafka writer for topic %s: %v", topic, err)
		}
	}
	return firstErr
}
