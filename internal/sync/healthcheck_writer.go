package sync

import (
	"context"
	"fmt"
)

// HealthCheckWriter wraps an EnvWriter and reports health based on write outcomes.
type HealthCheckWriter struct {
	inner      EnvWriter
	checker    *HealthChecker
	component  string
	failures   int
	threshold  int
}

// NewHealthCheckWriter creates a writer that tracks health via the provided checker.
// After threshold consecutive failures the component is marked Unhealthy.
func NewHealthCheckWriter(inner EnvWriter, checker *HealthChecker, component string, threshold int) *HealthCheckWriter {
	if threshold <= 0 {
		panic("healthcheck: threshold must be positive")
	}
	return &HealthCheckWriter{
		inner:     inner,
		checker:   checker,
		component: component,
		threshold: threshold,
	}
}

// Write delegates to the inner writer and updates the health status accordingly.
func (w *HealthCheckWriter) Write(ctx context.Context, key, value string) error {
	err := w.inner.Write(ctx, key, value)
	if err != nil {
		w.failures++
		status := Degraded
		if w.failures >= w.threshold {
			status = Unhealthy
		}
		w.checker.Record(w.component, status,
			fmt.Sprintf("%d consecutive failure(s): %v", w.failures, err))
		return err
	}
	w.failures = 0
	w.checker.Record(w.component, Healthy, "last write succeeded")
	return nil
}
