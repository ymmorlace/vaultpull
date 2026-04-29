package sync_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/your-org/vaultpull/internal/sync"
)

// mockEnvWriter captures written entries for assertions.
type mockEnvWriter struct {
	entries []string
	wantErr error
}

func (m *mockEnvWriter) Write(key, value string) error {
	if m.wantErr != nil {
		return m.wantErr
	}
	m.entries = append(m.entries, key+"="+value)
	return nil
}

func (m *mockEnvWriter) Close() error { return nil }

func TestTransformingWriter_AppliesTransform(t *testing.T) {
	inner := &mockEnvWriter{}
	tw := sync.NewTransformingWriter(inner, sync.NewTrimSpaceTransformer())

	if err := tw.Write("DB_HOST", "  localhost  "); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(inner.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(inner.entries))
	}
	if inner.entries[0] != "DB_HOST=localhost" {
		t.Errorf("got %q, want %q", inner.entries[0], "DB_HOST=localhost")
	}
}

func TestTransformingWriter_PropagatesInnerError(t *testing.T) {
	writeErr := errors.New("disk full")
	inner := &mockEnvWriter{wantErr: writeErr}
	tw := sync.NewTransformingWriter(inner, sync.NewTrimSpaceTransformer())

	err := tw.Write("KEY", "value")
	if !errors.Is(err, writeErr) {
		t.Errorf("expected disk full error, got %v", err)
	}
}

func TestTransformingWriter_QuotesSpecialValues(t *testing.T) {
	inner := &mockEnvWriter{}
	chain := sync.NewChainTransformer(
		sync.NewTrimSpaceTransformer(),
		sync.NewQuoteTransformer(),
	)
	tw := sync.NewTransformingWriter(inner, chain)

	if err := tw.Write("MSG", "  hello world  "); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(inner.entries[0], `"hello world"`) {
		t.Errorf("expected quoted value in entry, got %q", inner.entries[0])
	}
}

func TestTransformingWriter_Close(t *testing.T) {
	inner := &mockEnvWriter{}
	tw := sync.NewTransformingWriter(inner, sync.NewTrimSpaceTransformer())
	if err := tw.Close(); err != nil {
		t.Errorf("unexpected close error: %v", err)
	}
}
