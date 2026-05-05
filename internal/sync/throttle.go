package sync

import (
	"context"
	"fmt"
	"time"
)

// ThrottlePolicy defines how writes are throttled over time.
type ThrottlePolicy struct {
	// MinInterval is the minimum duration between successive writes.
	MinInterval time.Duration
}

// ThrottledWriter wraps an EnvWriter and enforces a minimum delay between writes.
type ThrottledWriter struct {
	inner    EnvWriter
	policy   ThrottlePolicy
	lastWrite time.Time
	now      func() time.Time
	sleep    func(context.Context, time.Duration) error
}

// NewThrottledWriter returns a ThrottledWriter that enforces policy between writes.
// Panics if inner is nil or MinInterval is zero.
func NewThrottledWriter(inner EnvWriter, policy ThrottlePolicy) *ThrottledWriter {
	if inner == nil {
		panic("throttle: inner writer must not be nil")
	}
	if policy.MinInterval <= 0 {
		panic("throttle: MinInterval must be positive")
	}
	return &ThrottledWriter{
		inner:  inner,
		policy: policy,
		now:    time.Now,
		sleep:  contextSleep,
	}
}

// Write enforces the throttle delay then delegates to the inner writer.
func (t *ThrottledWriter) Write(ctx context.Context, key, value string) error {
	if !t.lastWrite.IsZero() {
		elapsed := t.now().Sub(t.lastWrite)
		if elapsed < t.policy.MinInterval {
			wait := t.policy.MinInterval - elapsed
			if err := t.sleep(ctx, wait); err != nil {
				return fmt.Errorf("throttle: context cancelled while waiting: %w", err)
			}
		}
	}
	t.lastWrite = t.now()
	return t.inner.Write(ctx, key, value)
}

// contextSleep sleeps for d or until ctx is cancelled.
func contextSleep(ctx context.Context, d time.Duration) error {
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
