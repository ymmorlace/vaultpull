package sync_test

import (
	"context"
	"fmt"
	"testing"

	sync "github.com/yourusername/vaultpull/internal/sync"
)

// TestBufferedWriter_IntegrationWithDedup verifies that BufferedWriter flushes
// correctly into a DedupWriter, surfacing duplicate-key errors as expected.
func TestBufferedWriter_IntegrationWithDedup(t *testing.T) {
	cw := &captureWriter{}
	dedup := sync.NewDedupWriter(cw)
	bw := sync.NewBufferedWriter(dedup, 10)
	ctx := context.Background()

	keys := []string{"DB_HOST", "DB_PORT", "API_KEY"}
	for i, k := range keys {
		if err := bw.Write(ctx, k, fmt.Sprintf("value%d", i)); err != nil {
			t.Fatalf("unexpected write error for %s: %v", k, err)
		}
	}

	if err := bw.Flush(ctx); err != nil {
		t.Fatalf("unexpected flush error: %v", err)
	}

	if cw.count() != len(keys) {
		t.Fatalf("expected %d writes, got %d", len(keys), cw.count())
	}
}

// TestBufferedWriter_IntegrationWithDedup_DetectsDuplicate checks that a
// duplicate key written through BufferedWriter is caught by DedupWriter on flush.
func TestBufferedWriter_IntegrationWithDedup_DetectsDuplicate(t *testing.T) {
	cw := &captureWriter{}
	dedup := sync.NewDedupWriter(cw)
	bw := sync.NewBufferedWriter(dedup, 10)
	ctx := context.Background()

	_ = bw.Write(ctx, "TOKEN", "abc")
	_ = bw.Write(ctx, "TOKEN", "xyz") // duplicate

	if err := bw.Flush(ctx); err == nil {
		t.Fatal("expected duplicate-key error on flush, got nil")
	}
}
