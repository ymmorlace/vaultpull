package sync_test

import (
	"context"
	"strings"
	"testing"

	"github.com/your-org/vaultpull/internal/sync"
)

// multiCaptureWriter records all written key/value pairs.
type multiCaptureWriter struct {
	entries []struct{ key, value string }
}

func (m *multiCaptureWriter) Write(_ context.Context, key, value string) error {
	m.entries = append(m.entries, struct{ key, value string }{key, value})
	return nil
}

// TestEnvelopeWriter_IntegrationWithDedupWriter verifies that EnvelopeWriter
// composes cleanly with DedupWriter: duplicate wrapped keys are rejected.
func TestEnvelopeWriter_IntegrationWithDedupWriter(t *testing.T) {
	cap := &multiCaptureWriter{}
	dedup := sync.NewDedupWriter(cap)
	env := sync.NewEnvelopeWriter(dedup, "SVC", "", "__")

	ctx := context.Background()

	if err := env.Write(ctx, "PORT", "8080"); err != nil {
		t.Fatalf("first write failed: %v", err)
	}

	// Second write with the same key should be rejected by DedupWriter.
	err := env.Write(ctx, "PORT", "9090")
	if err == nil {
		t.Fatal("expected duplicate error, got nil")
	}
	if !strings.Contains(err.Error(), "SVC__PORT") {
		t.Errorf("error should mention wrapped key, got: %v", err)
	}

	if len(cap.entries) != 1 {
		t.Errorf("expected 1 entry recorded, got %d", len(cap.entries))
	}
}

// TestEnvelopeWriter_IntegrationWithBatchWriter verifies that EnvelopeWriter
// works inside a BatchWriter, wrapping all keys in the batch.
func TestEnvelopeWriter_IntegrationWithBatchWriter(t *testing.T) {
	cap := &multiCaptureWriter{}
	env := sync.NewEnvelopeWriter(cap, "BATCH", "END", "-")

	jobs := []sync.BatchJob{
		{Key: "ALPHA", Value: "1"},
		{Key: "BETA", Value: "2"},
		{Key: "GAMMA", Value: "3"},
	}

	bw := sync.NewBatchWriter(env)
	if err := bw.WriteAll(context.Background(), jobs); err != nil {
		t.Fatalf("WriteAll failed: %v", err)
	}

	expected := []string{"BATCH-ALPHA-END", "BATCH-BETA-END", "BATCH-GAMMA-END"}
	for i, e := range cap.entries {
		if e.key != expected[i] {
			t.Errorf("entry %d: got key %q, want %q", i, e.key, expected[i])
		}
	}
}
