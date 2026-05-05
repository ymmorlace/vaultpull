package sync

import "fmt"

// cbWriter wraps an EnvWriter and guards each write with a CircuitBreaker.
// If the circuit is open the write is skipped and ErrCircuitOpen is returned.
// Successful writes reset the failure counter; errors increment it.
type cbWriter struct {
	inner EnvWriter
	cb    *CircuitBreaker
}

// EnvWriter is the minimal interface expected by cbWriter and other decorators.
type EnvWriter interface {
	Write(key, value string) error
}

// NewCircuitBreakerWriter wraps inner with circuit-breaker protection.
func NewCircuitBreakerWriter(inner EnvWriter, cb *CircuitBreaker) EnvWriter {
	if cb == nil {
		cb = NewCircuitBreaker(3, 0)
	}
	return &cbWriter{inner: inner, cb: cb}
}

func (w *cbWriter) Write(key, value string) error {
	if err := w.cb.Allow(); err != nil {
		return fmt.Errorf("write skipped for key %q: %w", key, err)
	}
	if err := w.inner.Write(key, value); err != nil {
		w.cb.RecordFailure()
		return err
	}
	w.cb.RecordSuccess()
	return nil
}
