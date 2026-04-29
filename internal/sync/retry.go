package sync

import (
	"errors"
	"time"
)

// RetryPolicy defines how retries are performed on transient errors.
type RetryPolicy struct {
	MaxAttempts int
	Delay       time.Duration
	Multiplier  float64
}

// DefaultRetryPolicy returns a sensible default retry policy.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts: 3,
		Delay:       200 * time.Millisecond,
		Multiplier:  2.0,
	}
}

// IsTransient reports whether an error should trigger a retry.
// Callers may wrap errors with TransientError to signal retryability.
func IsTransient(err error) bool {
	var te *TransientError
	return errors.As(err, &te)
}

// TransientError wraps an error to mark it as transient (retryable).
type TransientError struct {
	Cause error
}

func (e *TransientError) Error() string {
	return "transient: " + e.Cause.Error()
}

func (e *TransientError) Unwrap() error { return e.Cause }

// Retry executes fn according to policy, retrying on transient errors.
// It returns the last error if all attempts are exhausted.
func Retry(policy RetryPolicy, fn func() error) error {
	if policy.MaxAttempts <= 0 {
		policy.MaxAttempts = 1
	}
	delay := policy.Delay
	var err error
	for attempt := 0; attempt < policy.MaxAttempts; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}
		if !IsTransient(err) {
			return err
		}
		if attempt < policy.MaxAttempts-1 {
			time.Sleep(delay)
			if policy.Multiplier > 0 {
				delay = time.Duration(float64(delay) * policy.Multiplier)
			}
		}
	}
	return err
}
