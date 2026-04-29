package sync

import (
	"errors"
	"testing"
	"time"
)

func TestRetry_SuccessOnFirstAttempt(t *testing.T) {
	calls := 0
	err := Retry(DefaultRetryPolicy(), func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestRetry_RetriesOnTransient(t *testing.T) {
	calls := 0
	policy := RetryPolicy{MaxAttempts: 3, Delay: time.Millisecond, Multiplier: 1.0}
	err := Retry(policy, func() error {
		calls++
		if calls < 3 {
			return &TransientError{Cause: errors.New("temporary")}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil after retries, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestRetry_StopsOnPermanentError(t *testing.T) {
	calls := 0
	permanent := errors.New("permanent failure")
	policy := RetryPolicy{MaxAttempts: 3, Delay: time.Millisecond, Multiplier: 1.0}
	err := Retry(policy, func() error {
		calls++
		return permanent
	})
	if !errors.Is(err, permanent) {
		t.Fatalf("expected permanent error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call for permanent error, got %d", calls)
	}
}

func TestRetry_ExhaustsAttempts(t *testing.T) {
	calls := 0
	policy := RetryPolicy{MaxAttempts: 3, Delay: time.Millisecond, Multiplier: 1.0}
	err := Retry(policy, func() error {
		calls++
		return &TransientError{Cause: errors.New("always fails")}
	})
	if err == nil {
		t.Fatal("expected error after exhausting attempts")
	}
	if !IsTransient(err) {
		t.Fatalf("expected transient error, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestIsTransient_WrappedError(t *testing.T) {
	base := errors.New("base")
	te := &TransientError{Cause: base}
	if !IsTransient(te) {
		t.Fatal("expected IsTransient to return true")
	}
	if IsTransient(base) {
		t.Fatal("expected IsTransient to return false for plain error")
	}
}

func TestTransientError_Unwrap(t *testing.T) {
	base := errors.New("root cause")
	te := &TransientError{Cause: base}
	if !errors.Is(te, base) {
		t.Fatal("expected errors.Is to find root cause through TransientError")
	}
}
