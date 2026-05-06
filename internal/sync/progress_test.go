package sync

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestProgressReporter_RecordWritten(t *testing.T) {
	var buf bytes.Buffer
	p := NewProgressReporter(&buf, 3)
	p.RecordWritten("FOO")

	if !strings.Contains(buf.String(), "wrote   FOO") {
		t.Errorf("expected wrote FOO in output, got: %s", buf.String())
	}
	w, sk, er := p.Counts()
	if w != 1 || sk != 0 || er != 0 {
		t.Errorf("unexpected counts: %d %d %d", w, sk, er)
	}
}

func TestProgressReporter_RecordSkipped(t *testing.T) {
	var buf bytes.Buffer
	p := NewProgressReporter(&buf, 0)
	p.RecordSkipped("BAR")

	if !strings.Contains(buf.String(), "skipped BAR") {
		t.Errorf("expected skipped BAR in output, got: %s", buf.String())
	}
	_, sk, _ := p.Counts()
	if sk != 1 {
		t.Errorf("expected 1 skipped, got %d", sk)
	}
}

func TestProgressReporter_RecordError(t *testing.T) {
	var buf bytes.Buffer
	p := NewProgressReporter(&buf, 0)
	p.RecordError("BAZ", errors.New("disk full"))

	if !strings.Contains(buf.String(), "error   BAZ: disk full") {
		t.Errorf("expected error line, got: %s", buf.String())
	}
	_, _, er := p.Counts()
	if er != 1 {
		t.Errorf("expected 1 error, got %d", er)
	}
}

func TestProgressReporter_Summary(t *testing.T) {
	var buf bytes.Buffer
	p := NewProgressReporter(&buf, 2)
	fixedNow := p.start.Add(250 * time.Millisecond)
	p.clock = func() time.Time { return fixedNow }
	p.RecordWritten("K1")
	p.RecordSkipped("K2")
	p.Summary()

	if !strings.Contains(buf.String(), "sync complete: 1 written, 1 skipped, 0 errors") {
		t.Errorf("unexpected summary: %s", buf.String())
	}
}

func TestProgressReporter_NilWriterDefaultsToStderr(t *testing.T) {
	// Should not panic when out is nil.
	p := NewProgressReporter(nil, 0)
	if p.out == nil {
		t.Error("expected non-nil writer")
	}
}

func TestProgressWriter_RecordsWritten(t *testing.T) {
	var buf bytes.Buffer
	p := NewProgressReporter(&buf, 1)
	inner := &captureWriter{}
	pw := NewProgressWriter(inner, p)

	if err := pw.Write(context.Background(), "KEY", "val"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	w, _, _ := p.Counts()
	if w != 1 {
		t.Errorf("expected 1 written, got %d", w)
	}
}

func TestProgressWriter_RecordsError(t *testing.T) {
	var buf bytes.Buffer
	p := NewProgressReporter(&buf, 1)
	inner := &failWriter{err: errors.New("boom")}
	pw := NewProgressWriter(inner, p)

	err := pw.Write(context.Background(), "KEY", "val")
	if err == nil {
		t.Fatal("expected error")
	}
	_, _, er := p.Counts()
	if er != 1 {
		t.Errorf("expected 1 error, got %d", er)
	}
}

// captureWriter records writes without error.
type captureWriter struct{ key, value string }

func (c *captureWriter) Write(_ context.Context, key, value string) error {
	c.key, c.value = key, value
	return nil
}

// failWriter always returns the configured error.
type failWriter struct{ err error }

func (f *failWriter) Write(_ context.Context, _, _ string) error { return f.err }
