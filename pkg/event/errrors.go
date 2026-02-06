package event

import "errors"

var (
	ErrRetryable = errors.New("retryable error")
	ErrFatal     = errors.New("fatal error")
)
