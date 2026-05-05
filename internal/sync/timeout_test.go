package sync_test

import (
	"context"
	"errors"
	"testing"
	"time"

	internalsync "github.com/example/vaultpull/internal/sync"
)

type slowWriter struct {
	delay time.Duration
	err   error
}

func (s *slowWriter) Write(ctx context.Context, _, _ string) error {
	select {
	case <-time.After(s.delay):
		return s.err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func TestTimeoutWriter_CompletesWithinDeadline(t *testing.T) {
	inner := &slowWriter{delay: 1 * time.Millisecond}
	w := internalsync.NewTimeoutWriter(inner, 100*time.Millisecond)

	if err := w.Write(context.Background(), "KEY", "val"); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestTimeoutWriter_ExceedsDeadline(t *testing.T) {
	inner := &slowWriter{delay: 200 * time.Millisecond}
	w := internalsync.NewTimeoutWriter(inner, 20*time.Millisecond)

	err := w.Write(context.Background(), "SLOW_KEY", "val")
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got: %v", err)
	}
}

func TestTimeoutWriter_PropagatesInnerError(t *testing.T) {
	sentinel := errors.New("inner failure")
	inner := &slowWriter{delay: 1 * time.Millisecond, err: sentinel}
	w := internalsync.NewTimeoutWriter(inner, 100*time.Millisecond)

	err := w.Write(context.Background(), "KEY", "val")
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got: %v", err)
	}
}

func TestTimeoutWriter_PanicOnZeroDuration(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero duration")
		}
	}()
	internalsync.NewTimeoutWriter(&slowWriter{}, 0)
}

func TestTimeoutWriter_RespectsParentCancellation(t *testing.T) {
	inner := &slowWriter{delay: 500 * time.Millisecond}
	w := internalsync.NewTimeoutWriter(inner, 2*time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err := w.Write(ctx, "KEY", "val")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected Canceled, got: %v", err)
	}
}
