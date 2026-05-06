package sync_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/your-org/vaultpull/internal/sync"
)

func TestDryRunWriter_RecordsEntries(t *testing.T) {
	var buf bytes.Buffer
	w := sync.NewDryRunWriter(&buf)
	ctx := context.Background()

	if err := w.Write(ctx, "DB_HOST", "localhost"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := w.Write(ctx, "DB_PORT", "5432"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries := w.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Key != "DB_HOST" || entries[0].Value != "localhost" {
		t.Errorf("unexpected first entry: %+v", entries[0])
	}
	if entries[1].Key != "DB_PORT" || entries[1].Value != "5432" {
		t.Errorf("unexpected second entry: %+v", entries[1])
	}
}

func TestDryRunWriter_PrintsToWriter(t *testing.T) {
	var buf bytes.Buffer
	w := sync.NewDryRunWriter(&buf)

	_ = w.Write(context.Background(), "SECRET_KEY", "abc123")

	output := buf.String()
	if !strings.Contains(output, "[dry-run]") {
		t.Errorf("expected [dry-run] prefix in output, got: %q", output)
	}
	if !strings.Contains(output, "SECRET_KEY") {
		t.Errorf("expected key in output, got: %q", output)
	}
}

func TestDryRunWriter_DefaultsToStdout(t *testing.T) {
	// Should not panic when out is nil.
	w := sync.NewDryRunWriter(nil)
	if w == nil {
		t.Fatal("expected non-nil writer")
	}
}

func TestDryRunWriter_Reset_ClearsEntries(t *testing.T) {
	var buf bytes.Buffer
	w := sync.NewDryRunWriter(&buf)
	ctx := context.Background()

	_ = w.Write(ctx, "FOO", "bar")
	if len(w.Entries()) != 1 {
		t.Fatalf("expected 1 entry before reset")
	}

	w.Reset()
	if len(w.Entries()) != 0 {
		t.Errorf("expected 0 entries after reset, got %d", len(w.Entries()))
	}
}

func TestDryRunWriter_EntriesReturnsCopy(t *testing.T) {
	var buf bytes.Buffer
	w := sync.NewDryRunWriter(&buf)
	_ = w.Write(context.Background(), "K", "V")

	e1 := w.Entries()
	e1[0].Key = "MUTATED"

	e2 := w.Entries()
	if e2[0].Key == "MUTATED" {
		t.Error("Entries() should return a copy, not a reference to internal state")
	}
}
