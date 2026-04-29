package sync_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/your-org/vaultpull/internal/sync"
)

func TestAuditLog_RecordOK(t *testing.T) {
	var buf bytes.Buffer
	al := sync.NewAuditLog(&buf)

	if err := al.RecordOK("secret/app", "DB_PASS"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	line := buf.String()
	for _, want := range []string{"secret/app", "DB_PASS", "ok"} {
		if !strings.Contains(line, want) {
			t.Errorf("expected %q in output %q", want, line)
		}
	}
}

func TestAuditLog_RecordSkipped(t *testing.T) {
	var buf bytes.Buffer
	al := sync.NewAuditLog(&buf)

	if err := al.RecordSkipped("secret/app", "OLD_KEY", "duplicate"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	line := buf.String()
	if !strings.Contains(line, "skipped") {
		t.Errorf("expected 'skipped' in %q", line)
	}
	if !strings.Contains(line, "duplicate") {
		t.Errorf("expected message 'duplicate' in %q", line)
	}
}

func TestAuditLog_RecordError(t *testing.T) {
	var buf bytes.Buffer
	al := sync.NewAuditLog(&buf)

	if err := al.RecordError("secret/app", "API_KEY", errors.New("vault timeout")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	line := buf.String()
	if !strings.Contains(line, "error") {
		t.Errorf("expected 'error' in %q", line)
	}
	if !strings.Contains(line, "vault timeout") {
		t.Errorf("expected error message in %q", line)
	}
}

func TestAuditLog_NilWriter_Discards(t *testing.T) {
	al := sync.NewAuditLog(nil)
	if err := al.RecordOK("secret/app", "KEY"); err != nil {
		t.Fatalf("expected no error with discard writer, got: %v", err)
	}
}

func TestAuditLog_TimestampIsSet(t *testing.T) {
	var buf bytes.Buffer
	al := sync.NewAuditLog(&buf)

	before := time.Now().UTC().Truncate(time.Second)
	_ = al.RecordOK("secret/app", "KEY")

	line := buf.String()
	year := before.Format("2006")
	if !strings.Contains(line, year) {
		t.Errorf("expected year %s in timestamp output %q", year, line)
	}
}
