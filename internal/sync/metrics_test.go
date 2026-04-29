package sync_test

import (
	"errors"
	"testing"
	"time"

	"github.com/user/vaultpull/internal/sync"
)

// stubWriter is a minimal EnvWriter for testing.
type stubMetricsWriter struct {
	err error
	calls []string
}

func (s *stubMetricsWriter) Write(key, _ string) error {
	s.calls = append(s.calls, key)
	return s.err
}

func TestSyncMetrics_RecordWritten(t *testing.T) {
	m := sync.NewSyncMetricsExported()
	m.RecordWritten()
	m.RecordWritten()
	if m.Written != 2 {
		t.Fatalf("expected Written=2, got %d", m.Written)
	}
}

func TestSyncMetrics_RecordSkipped(t *testing.T) {
	m := sync.NewSyncMetricsExported()
	m.RecordSkipped()
	if m.Skipped != 1 {
		t.Fatalf("expected Skipped=1, got %d", m.Skipped)
	}
}

func TestSyncMetrics_RecordError(t *testing.T) {
	m := sync.NewSyncMetricsExported()
	m.RecordError()
	m.RecordError()
	m.RecordError()
	if m.Errors != 3 {
		t.Fatalf("expected Errors=3, got %d", m.Errors)
	}
}

func TestSyncMetrics_Duration(t *testing.T) {
	m := sync.NewSyncMetricsExported()
	time.Sleep(5 * time.Millisecond)
	m.Finish()
	if m.Duration() < 5*time.Millisecond {
		t.Fatalf("expected duration >= 5ms, got %v", m.Duration())
	}
}

func TestSyncMetrics_Summary(t *testing.T) {
	m := sync.NewSyncMetricsExported()
	m.RecordWritten()
	m.RecordSkipped()
	m.RecordError()
	m.Finish()
	s := m.Summary()
	if s.Written != 1 || s.Skipped != 1 || s.Errors != 1 {
		t.Fatalf("unexpected summary: %+v", s)
	}
}

func TestMetricsWriter_RecordsWritten(t *testing.T) {
	stub := &stubMetricsWriter{}
	m := sync.NewSyncMetricsExported()
	w := sync.NewMetricsWriterExported(stub, m)

	if err := w.Write("KEY", "val"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Written != 1 || m.Errors != 0 {
		t.Fatalf("expected Written=1 Errors=0, got Written=%d Errors=%d", m.Written, m.Errors)
	}
}

func TestMetricsWriter_RecordsError(t *testing.T) {
	stub := &stubMetricsWriter{err: errors.New("disk full")}
	m := sync.NewSyncMetricsExported()
	w := sync.NewMetricsWriterExported(stub, m)

	err := w.Write("KEY", "val")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if m.Errors != 1 || m.Written != 0 {
		t.Fatalf("expected Errors=1 Written=0, got Errors=%d Written=%d", m.Errors, m.Written)
	}
}
