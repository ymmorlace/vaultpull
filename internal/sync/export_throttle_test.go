package sync

import (
	"context"
	"time"
)

// NewThrottledWriterExported exposes constructor for black-box tests.
func NewThrottledWriterExported(inner EnvWriter, interval time.Duration) *ThrottledWriter {
	return NewThrottledWriter(inner, ThrottlePolicy{MinInterval: interval})
}

// OverrideThrottleClock replaces the clock and sleep functions for deterministic tests.
func OverrideThrottleClock(
	tw *ThrottledWriter,
	nowFn func() time.Time,
	sleepFn func(context.Context, time.Duration) error,
) {
	tw.now = nowFn
	tw.sleep = sleepFn
}
