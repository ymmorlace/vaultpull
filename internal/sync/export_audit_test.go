package sync

import "io"

// NewAuditLogExported exposes NewAuditLog for black-box tests in sync_test package.
func NewAuditLogExported(w io.Writer) *AuditLog {
	return NewAuditLog(w)
}
