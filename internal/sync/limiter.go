package sync

import (
	"context"
	"errors"
	"fmt"
)

// ErrLimitExceeded is returned when the write limit has been reached.
var ErrLimitExceeded = errors.New("write limit exceeded")

// LimitWriter wraps an EnvWriter and enforces a maximum number of Write calls.
// Once the limit is reached, subsequent writes return ErrLimitExceeded.
type LimitWriter struct {
	inner EnvWriter
	limit int
	count int
}

// NewLimitWriter creates a LimitWriter that allows at most limit writes
// through to inner. Panics if inner is nil or limit is not positive.
func NewLimitWriter(inner EnvWriter, limit int) *LimitWriter {
	if inner == nil {
		panic("limiter: inner writer must not be nil")
	}
	if limit <= 0 {
		panic(fmt.Sprintf("limiter: limit must be positive, got %d", limit))
	}
	return &LimitWriter{inner: inner, limit: limit}
}

// Write forwards the entry to the inner writer if the limit has not been
// reached. Returns ErrLimitExceeded once the cap is hit.
func (l *LimitWriter) Write(ctx context.Context, key, value string) error {
	if l.count >= l.limit {
		return fmt.Errorf("%w: limit is %d", ErrLimitExceeded, l.limit)
	}
	if err := l.inner.Write(ctx, key, value); err != nil {
		return err
	}
	l.count++
	return nil
}

// Count returns the number of successful writes recorded so far.
func (l *LimitWriter) Count() int {
	return l.count
}

// Reset resets the write counter, allowing the writer to be reused.
func (l *LimitWriter) Reset() {
	l.count = 0
}
