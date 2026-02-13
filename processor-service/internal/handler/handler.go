package handler

import (
	"context"
)

func HandleEvent(ctx context.Context, key string, value []byte) error {
	// TODO: business logic
	// return ErrRetryable for transient failures
	// return ErrFatal for permanent failures

	return nil
}
