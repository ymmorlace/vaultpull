package sync_test

import (
	"context"
	"testing"

	"github.com/your-org/vaultpull/internal/sync"
)

func TestSplitter_IntegrationWithDedupWriter(t *testing.T) {
	a := &sync.CaptureWriterExported{}
	b := &sync.CaptureWriterExported{}

	dedupA := sync.NewDedupWriter(a)
	dedupB := sync.NewDedupWriter(b)

	router := func(key string) int {
		if len(key) > 0 && key[0] == 'A' {
			return 0
		}
		return 1
	}

	s := sync.NewSplitterExported(router, dedupA, dedupB)
	ctx := context.Background()

	pairs := []struct{ k, v string }{
		{"ALPHA", "1"},
		{"BETA", "2"},
		{"ALPHA", "3"}, // duplicate – should error from dedupA
		{"GAMMA", "4"},
	}

	var errCount int
	for _, p := range pairs {
		if err := s.Write(ctx, p.k, p.v); err != nil {
			errCount++
		}
	}

	if errCount != 1 {
		t.Fatalf("expected 1 dedup error, got %d", errCount)
	}

	if s.Len() != 2 {
		t.Fatalf("expected 2 backing writers, got %d", s.Len())
	}
}
