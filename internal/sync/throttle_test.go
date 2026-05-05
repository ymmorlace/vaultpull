package sync

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestThrottledWriter_PanicOnNilInner(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner writer")
		}
	}()
	NewThrottledWriter(nil, ThrottlePolicy{MinInterval: 10 * time.Millisecond})
}

func TestThrottledWriter_PanicOnZeroInterval(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero MinInterval")
		}
	}()
	NewThrottledWriter(&captureWriter{}, ThrottlePolicy{})
}

func TestThrottledWriter_FirstWriteNoDelay(t *testing.T) {
	cap := &captureWriter{}
	tw := NewThrottledWriter(cap, ThrottlePolicy{MinInterval: 50 * time.Millisecond})

	start := time.Now()
	if err := tw.Write(context.Background(), "KEY", "val"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 20*time.Millisecond {
		t.Fatalf("first write should not be delayed, took %v", elapsed)
	}
	if cap.lastKey != "KEY" {
		t.Fatalf("expected KEY, got %s", cap.lastKey)
	}
}

func TestThrottledWriter_EnforcesMinInterval(t *testing.T) {
	cap := &captureWriter{}
	interval := 40 * time.Millisecond
	tw := NewThrottledWriter(cap, ThrottlePolicy{MinInterval: interval})

	_ = tw.Write(context.Background(), "A", "1")
	start := time.Now()
	_ = tw.Write(context.Background(), "B", "2")
	elapsed := time.Since(start)

	if elapsed < interval-5*time.Millisecond {
		t.Fatalf("expected throttle delay >= %v, got %v", interval, elapsed)
	}
}

func TestThrottledWriter_ContextCancelledDuringWait(t *testing.T) {
	cap := &captureWriter{}
	tw := NewThrottledWriter(cap, ThrottlePolicy{MinInterval: 200 * time.Millisecond})

	_ = tw.Write(context.Background(), "FIRST", "v")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := tw.Write(ctx, "SECOND", "v")
	if err == nil {
		t.Fatal("expected error when context is cancelled")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestThrottledWriter_PropagatesInnerError(t *testing.T) {
	errWriter := &errorWriter{err: errors.New("disk full")}
	tw := NewThrottledWriter(errWriter, ThrottlePolicy{MinInterval: 1 * time.Millisecond})

	err := tw.Write(context.Background(), "K", "V")
	if err == nil || err.Error() != "disk full" {
		t.Fatalf("expected inner error, got %v", err)
	}
}

// captureWriter records the last key/value written.
type captureWriter struct {
	lastKey   string
	lastValue string
}

func (c *captureWriter) Write(_ context.Context, key, value string) error {
	c.lastKey = key
	c.lastValue = value
	return nil
}

// errorWriter always returns the configured error.
type errorWriter struct{ err error }

func (e *errorWriter) Write(_ context.Context, _, _ string) error { return e.err }
