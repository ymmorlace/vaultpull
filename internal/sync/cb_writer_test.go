package sync_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	internalsync "github.com/example/vaultpull/internal/sync"
)

// stubWriter records calls and can be configured to return an error.
type stubWriter struct {
	calls  []string
	errOn string
}

func (s *stubWriter) Write(key, value string) error {
	if key == s.errOn {
		return fmt.Errorf("write error for %s", key)
	}
	s.calls = append(s.calls, key+"="+value)
	return nil
}

func TestCBWriter_PassesThroughWhenClosed(t *testing.T) {
	inner := &stubWriter{}
	cb := internalsync.NewCircuitBreaker(3, time.Minute)
	w := internalsync.NewCircuitBreakerWriter(inner, cb)

	if err := w.Write("FOO", "bar"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(inner.calls) != 1 || inner.calls[0] != "FOO=bar" {
		t.Fatalf("expected write to be forwarded, got %v", inner.calls)
	}
}

func TestCBWriter_RecordsFailureOnError(t *testing.T) {
	inner := &stubWriter{errOn: "BAD"}
	cb := internalsync.NewCircuitBreaker(2, time.Minute)
	w := internalsync.NewCircuitBreakerWriter(inner, cb)

	_ = w.Write("BAD", "v")
	_ = w.Write("BAD", "v")

	if cb.State() != "open" {
		t.Fatalf("expected circuit open after threshold failures, got %s", cb.State())
	}
}

func TestCBWriter_BlocksWhenOpen(t *testing.T) {
	inner := &stubWriter{}
	cb := internalsync.NewCircuitBreaker(1, time.Minute)
	cb.RecordFailure() // force open
	w := internalsync.NewCircuitBreakerWriter(inner, cb)

	err := w.Write("KEY", "val")
	if !errors.Is(err, internalsync.ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
	if len(inner.calls) != 0 {
		t.Fatal("inner writer should not be called when circuit is open")
	}
}

func TestCBWriter_ClosesAfterSuccessfulRecovery(t *testing.T) {
	inner := &stubWriter{}
	cb := internalsync.NewCircuitBreaker(1, 10*time.Millisecond)
	cb.RecordFailure()
	time.Sleep(20 * time.Millisecond)

	w := internalsync.NewCircuitBreakerWriter(inner, cb)
	if err := w.Write("RECOVER", "ok"); err != nil {
		t.Fatalf("expected recovery write to succeed, got %v", err)
	}
	if cb.State() != "closed" {
		t.Fatalf("expected circuit closed after successful recovery, got %s", cb.State())
	}
}
