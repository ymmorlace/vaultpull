package sync

// Exported helpers for healthcheck white-box testing.

// NewHealthCheckerExported exposes NewHealthChecker for external test packages.
func NewHealthCheckerExported() *HealthChecker {
	return NewHealthChecker()
}

// NewHealthCheckWriterExported exposes NewHealthCheckWriter for external test packages.
func NewHealthCheckWriterExported(inner EnvWriter, checker *HealthChecker, component string, threshold int) *HealthCheckWriter {
	return NewHealthCheckWriter(inner, checker, component, threshold)
}
