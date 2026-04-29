package sync_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/your-org/vaultpull/internal/sync"
)

type stubWriter struct {
	calls []string
	errOn string
}

func (s *stubWriter) Write(path, key, value string) error {
	s.calls = append(s.calls, key)
	if key == s.errOn {
		return errors.New("write failed")
	}
	return nil
}

func TestAuditingWriter_RecordsOK(t *testing.T) {
	var buf bytes.Buffer
	log := sync.NewAuditLog(&buf)
	stub := &stubWriter{}
	aw := sync.NewAuditingWriter(stub, log)

	if err := aw.Write("secret/app", "DB_URL", "postgres://localhost"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "ok") {
		t.Errorf("expected 'ok' in audit log, got: %s", buf.String())
	}
	if len(stub.calls) != 1 || stub.calls[0] != "DB_URL" {
		t.Errorf("expected inner writer called with DB_URL")
	}
}

func TestAuditingWriter_RecordsError(t *testing.T) {
	var buf bytes.Buffer
	log := sync.NewAuditLog(&buf)
	stub := &stubWriter{errOn: "BAD_KEY"}
	aw := sync.NewAuditingWriter(stub, log)

	err := aw.Write("secret/app", "BAD_KEY", "value")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	output := buf.String()
	if !strings.Contains(output, "error") {
		t.Errorf("expected 'error' in audit log, got: %s", output)
	}
	if !strings.Contains(output, "write failed") {
		t.Errorf("expected error message in audit log, got: %s", output)
	}
}

func TestAuditingWriter_WrapsError(t *testing.T) {
	var buf bytes.Buffer
	log := sync.NewAuditLog(&buf)
	stub := &stubWriter{errOn: "KEY"}
	aw := sync.NewAuditingWriter(stub, log)

	err := aw.Write("secret/app", "KEY", "val")
	if err == nil {
		t.Fatal("expected wrapped error")
	}
	if !strings.Contains(err.Error(), "auditing writer") {
		t.Errorf("expected wrapped error message, got: %v", err)
	}
}
