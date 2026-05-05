package sync

import "time"

// NewConstantBackoffExported exposes ConstantBackoff for use in external test packages.
func NewConstantBackoffExported(interval time.Duration) BackoffPolicy {
	return &ConstantBackoff{Interval: interval}
}

// NewExponentialBackoffExported exposes ExponentialBackoff for use in external test packages.
func NewExponentialBackoffExported(base, max time.Duration, multiplier float64, jitter bool) BackoffPolicy {
	return &ExponentialBackoff{
		Base:       base,
		Max:        max,
		Multiplier: multiplier,
		Jitter:     jitter,
	}
}
