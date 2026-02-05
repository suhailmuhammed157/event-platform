package handler

import (
	"context"
	"log"
)

func HandleEvent(ctx context.Context, key string, value []byte) error {
	// Simulate processing
	log.Printf("processing event key=%s size=%d", key, len(value))
	return nil
}
