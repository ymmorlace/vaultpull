package sync

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrCircuitOpen is returned when the circuit breaker is in the open state.
var ErrCircuitOpen = errors.New("circuit breaker is open")

type cbState int

const (
	cbClosed cbState = iota
	cbOpen
	cbHalfOpen
)

// CircuitBreaker prevents repeated calls to a failing dependency.
type CircuitBreaker struct {
	mu           sync.Mutex
	state        cbState
	failures     int
	threshold    int
	resetTimeout time.Duration
	openedAt     time.Time
}

// NewCircuitBreaker creates a CircuitBreaker that opens after threshold
// consecutive failures and attempts recovery after resetTimeout.
func NewCircuitBreaker(threshold int, resetTimeout time.Duration) *CircuitBreaker {
	if threshold <= 0 {
		threshold = 3
	}
	if resetTimeout <= 0 {
		resetTimeout = 30 * time.Second
	}
	return &CircuitBreaker{
		threshold:    threshold,
		resetTimeout: resetTimeout,
	}
}

// Allow reports whether a call should be attempted.
// Returns ErrCircuitOpen when the circuit is open and the reset timeout has
// not yet elapsed.
func (cb *CircuitBreaker) Allow() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case cbOpen:
		if time.Since(cb.openedAt) >= cb.resetTimeout {
			cb.state = cbHalfOpen
			return nil
		}
		return fmt.Errorf("%w: retry after %s",
			ErrCircuitOpen, cb.resetTimeout-time.Since(cb.openedAt).Round(time.Second))
	default:
		return nil
	}
}

// RecordSuccess resets the failure counter and closes the circuit.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = cbClosed
}

// RecordFailure increments the failure counter and opens the circuit when
// the threshold is reached.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	if cb.failures >= cb.threshold {
		cb.state = cbOpen
		cb.openedAt = time.Now()
	}
}

// State returns a human-readable description of the current state.
func (cb *CircuitBreaker) State() string {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case cbOpen:
		return "open"
	case cbHalfOpen:
		return "half-open"
	default:
		return "closed"
	}
}
