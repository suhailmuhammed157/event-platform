// pkg/kafka/handler.go
package kafka

import "context"

type JobHandler interface {
	Handle(ctx context.Context, key string, value []byte) error
}

type HandlerFunc func(ctx context.Context, key string, value []byte) error

func (f HandlerFunc) Handle(ctx context.Context, key string, value []byte) error {
	return f(ctx, key, value)
}
