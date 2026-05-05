package sync_test

import (
	"errors"
	"testing"
	"time"

	internalsync "github.com/example/vaultpull/internal/sync"
)

func TestCircuitBreaker_InitiallyClosed(t *testing.T) {
	cb := internalsync.NewCircuitBreaker(3, time.Second)
	if err := cb.Allow(); err != nil {
		t.Fatalf("expected circuit closed, got error: %v", err)
	}
	if cb.State() != "closed" {
		t.Fatalf("expected state closed, got %s", cb.State())
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	cb := internalsync.NewCircuitBreaker(3, time.Minute)
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	if cb.State() != "open" {
		t.Fatalf("expected state open, got %s", cb.State())
	}
	err := cb.Allow()
	if !errors.Is(err, internalsync.ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitBreaker_HalfOpenAfterTimeout(t *testing.T) {
	cb := internalsync.NewCircuitBreaker(1, 10*time.Millisecond)
	cb.RecordFailure()
	time.Sleep(20 * time.Millisecond)

	if err := cb.Allow(); err != nil {
		t.Fatalf("expected half-open to allow, got: %v", err)
	}
	if cb.State() != "half-open" {
		t.Fatalf("expected state half-open, got %s", cb.State())
	}
}

func TestCircuitBreaker_ClosesOnSuccess(t *testing.T) {
	cb := internalsync.NewCircuitBreaker(1, 10*time.Millisecond)
	cb.RecordFailure()
	time.Sleep(20 * time.Millisecond)
	_ = cb.Allow() // transition to half-open
	cb.RecordSuccess()

	if cb.State() != "closed" {
		t.Fatalf("expected state closed after success, got %s", cb.State())
	}
	if err := cb.Allow(); err != nil {
		t.Fatalf("expected no error after close, got %v", err)
	}
}

func TestCircuitBreaker_SuccessResetsFailureCount(t *testing.T) {
	cb := internalsync.NewCircuitBreaker(3, time.Minute)
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordSuccess()
	cb.RecordFailure() // only 1 failure after reset; should not open

	if cb.State() != "closed" {
		t.Fatalf("expected circuit still closed, got %s", cb.State())
	}
}

func TestCircuitBreaker_DefaultThreshold(t *testing.T) {
	// threshold=0 should default to 3
	cb := internalsync.NewCircuitBreaker(0, time.Minute)
	for i := 0; i < 2; i++ {
		cb.RecordFailure()
	}
	if cb.State() != "closed" {
		t.Fatalf("expected closed after 2 failures with default threshold 3")
	}
	cb.RecordFailure()
	if cb.State() != "open" {
		t.Fatalf("expected open after 3 failures")
	}
}
