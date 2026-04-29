package sync

import (
	"fmt"
	"io"
	"os"
	"time"
)

// AuditEntry records a single secret sync operation.
type AuditEntry struct {
	Timestamp time.Time
	Path      string
	Key       string
	Status    string // "ok", "skipped", "error"
	Message   string
}

// AuditLog writes structured audit entries to a writer.
type AuditLog struct {
	w io.Writer
}

// NewAuditLog creates an AuditLog writing to w.
// Pass nil to discard all output.
func NewAuditLog(w io.Writer) *AuditLog {
	if w == nil {
		w = io.Discard
	}
	return &AuditLog{w: w}
}

// NewFileAuditLog opens (or creates) a file at path for appending and returns
// an AuditLog backed by it. The caller must close the returned *os.File.
func NewFileAuditLog(path string) (*AuditLog, *os.File, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, nil, fmt.Errorf("audit: open %s: %w", path, err)
	}
	return NewAuditLog(f), f, nil
}

// Record writes a single entry to the log.
func (a *AuditLog) Record(e AuditEntry) error {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}
	line := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\n",
		e.Timestamp.Format(time.RFC3339),
		e.Path,
		e.Key,
		e.Status,
		e.Message,
	)
	_, err := fmt.Fprint(a.w, line)
	return err
}

// RecordOK is a convenience wrapper for successful syncs.
func (a *AuditLog) RecordOK(path, key string) error {
	return a.Record(AuditEntry{Path: path, Key: key, Status: "ok"})
}

// RecordSkipped is a convenience wrapper for skipped keys.
func (a *AuditLog) RecordSkipped(path, key, reason string) error {
	return a.Record(AuditEntry{Path: path, Key: key, Status: "skipped", Message: reason})
}

// RecordError is a convenience wrapper for failed syncs.
func (a *AuditLog) RecordError(path, key string, err error) error {
	return a.Record(AuditEntry{Path: path, Key: key, Status: "error", Message: err.Error()})
}
