package sync

import (
	"math"
	"math/rand"
	"time"
)

// BackoffPolicy defines how delays are computed between retry attempts.
type BackoffPolicy interface {
	// Delay returns the duration to wait before the nth attempt (0-indexed).
	Delay(attempt int) time.Duration
}

// ExponentialBackoff implements truncated exponential backoff with optional jitter.
type ExponentialBackoff struct {
	// Base is the initial delay for attempt 0.
	Base time.Duration
	// Max is the upper bound on the computed delay.
	Max time.Duration
	// Multiplier is the growth factor per attempt (default 2.0).
	Multiplier float64
	// Jitter, when true, adds a random fraction of the computed delay.
	Jitter bool
}

// DefaultExponentialBackoff returns a sensible default backoff policy.
func DefaultExponentialBackoff() *ExponentialBackoff {
	return &ExponentialBackoff{
		Base:       200 * time.Millisecond,
		Max:        30 * time.Second,
		Multiplier: 2.0,
		Jitter:     true,
	}
}

// Delay computes the wait duration for the given attempt index.
func (e *ExponentialBackoff) Delay(attempt int) time.Duration {
	mul := e.Multiplier
	if mul <= 0 {
		mul = 2.0
	}

	scaled := float64(e.Base) * math.Pow(mul, float64(attempt))
	if scaled > float64(e.Max) {
		scaled = float64(e.Max)
	}

	d := time.Duration(scaled)

	if e.Jitter && d > 0 {
		// Add up to 50% random jitter.
		jitter := time.Duration(rand.Int63n(int64(d) / 2)) //nolint:gosec
		d += jitter
	}

	return d
}

// ConstantBackoff waits the same duration between every attempt.
type ConstantBackoff struct {
	Interval time.Duration
}

// Delay returns the constant interval regardless of attempt number.
func (c *ConstantBackoff) Delay(_ int) time.Duration {
	return c.Interval
}
