package sync

import (
	"context"
	"fmt"
	"time"
)

// TimeoutWriter wraps an EnvWriter and enforces a per-write deadline.
// If the underlying Write does not complete within the configured duration,
// the context is cancelled and an error is returned.
type TimeoutWriter struct {
	inner   EnvWriter
	timeout time.Duration
}

// NewTimeoutWriter creates a TimeoutWriter that cancels writes exceeding d.
func NewTimeoutWriter(inner EnvWriter, d time.Duration) *TimeoutWriter {
	if d <= 0 {
		panic("timeout: duration must be positive")
	}
	return &TimeoutWriter{inner: inner, timeout: d}
}

// Write delegates to the inner writer within a bounded context.
// It returns an error if the deadline is exceeded before the write completes.
func (t *TimeoutWriter) Write(ctx context.Context, key, value string) error {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	type result struct {
		err error
	}

	ch := make(chan result, 1)
	go func() {
		ch <- result{err: t.inner.Write(ctx, key, value)}
	}()

	select {
	case res := <-ch:
		return res.err
	case <-ctx.Done():
		return fmt.Errorf("timeout: write for key %q exceeded %s: %w", key, t.timeout, ctx.Err())
	}
}
