package sync

import "fmt"

// AuditingWriter wraps an EnvWriter and records every write attempt
// to an AuditLog.
type AuditingWriter struct {
	inner EnvWriter
	log   *AuditLog
}

// EnvWriter is the minimal interface for writing env key/value pairs.
type EnvWriter interface {
	Write(path, key, value string) error
}

// NewAuditingWriter wraps inner with audit logging.
func NewAuditingWriter(inner EnvWriter, log *AuditLog) *AuditingWriter {
	return &AuditingWriter{inner: inner, log: log}
}

// Write delegates to the inner writer and records the outcome.
func (a *AuditingWriter) Write(path, key, value string) error {
	err := a.inner.Write(path, key, value)
	if err != nil {
		_ = a.log.RecordError(path, key, err)
		return fmt.Errorf("auditing writer: %w", err)
	}
	_ = a.log.RecordOK(path, key)
	return nil
}
