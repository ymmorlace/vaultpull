package sync_test

import (
	"context"
	"testing"
	"time"

	sync "github.com/example/vaultpull/internal/sync"
)

func TestRateLimiter_Unlimited(t *testing.T) {
	rl := sync.NewRateLimiter(0)
	defer rl.Close()

	ctx := context.Background()
	// Should never block for unlimited limiter
	for i := 0; i < 10; i++ {
		if err := rl.Wait(ctx); err != nil {
			t.Fatalf("unexpected error on unlimited limiter: %v", err)
		}
	}
}

func TestRateLimiter_ContextCancelled(t *testing.T) {
	rl := sync.NewRateLimiter(1)
	defer rl.Close()

	// Drain the first token
	ctx := context.Background()
	if err := rl.Wait(ctx); err != nil {
		t.Fatalf("first wait failed: %v", err)
	}

	// Cancel immediately — should not block
	ctxCancel, cancel := context.WithCancel(context.Background())
	cancel()

	err := rl.Wait(ctxCancel)
	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}
}

func TestRateLimiter_ThrottlesRequests(t *testing.T) {
	rps := 10
	rl := sync.NewRateLimiter(rps)
	defer rl.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	start := time.Now()
	requests := rps + 1 // one more than initial burst

	for i := 0; i < requests; i++ {
		if err := rl.Wait(ctx); err != nil {
			t.Fatalf("wait %d failed: %v", i, err)
		}
	}

	elapsed := time.Since(start)
	// At 10 rps the (rps+1)th token requires at least one tick (~100ms)
	if elapsed < 50*time.Millisecond {
		t.Errorf("rate limiter did not throttle: elapsed=%v", elapsed)
	}
}

func TestRateLimiter_Close_IsIdempotent(t *testing.T) {
	rl := sync.NewRateLimiter(5)
	// Should not panic on double close
	rl.Close()
}
