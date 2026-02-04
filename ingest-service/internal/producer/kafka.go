package producer

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(broker, topic string) *KafkaProducer {
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(broker),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			BatchTimeout: 10 * time.Millisecond, // small batch for performance
		},
	}
}

func (p *KafkaProducer) Publish(ctx context.Context, key string, value []byte) error {
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: value,
	})
}

func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}
