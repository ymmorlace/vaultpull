package sync

import (
	"context"
	"time"
)

// RateLimiter controls the rate of secret reads to avoid overwhelming Vault.
type RateLimiter struct {
	ticker  *time.Ticker
	done    chan struct{}
	tokens  chan struct{}
}

// NewRateLimiter creates a RateLimiter that allows up to rps requests per second.
// A rps of 0 disables rate limiting (unlimited throughput).
func NewRateLimiter(rps int) *RateLimiter {
	if rps <= 0 {
		return &RateLimiter{}
	}

	rl := &RateLimiter{
		ticker: time.NewTicker(time.Second / time.Duration(rps)),
		done:   make(chan struct{}),
		tokens: make(chan struct{}, rps),
	}

	go func() {
		defer rl.ticker.Stop()
		for {
			select {
			case <-rl.ticker.C:
				select {
				case rl.tokens <- struct{}{}:
				default:
				}
			case <-rl.done:
				return
			}
		}
	}()

	return rl
}

// Wait blocks until a token is available or ctx is cancelled.
// Returns ctx.Err() if the context is done before a token is acquired.
func (rl *RateLimiter) Wait(ctx context.Context) error {
	if rl.tokens == nil {
		// unlimited
		return ctx.Err()
	}
	select {
	case <-rl.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Close shuts down the background goroutine.
func (rl *RateLimiter) Close() {
	if rl.done != nil {
		close(rl.done)
	}
}
