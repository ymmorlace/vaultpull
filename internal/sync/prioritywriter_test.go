package sync_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultpull/internal/sync"
)

// captureWriter records writes in order.
type captureWriter struct {
	keys   []string
	values []string
	failOn string
}

func (c *captureWriter) Write(_ context.Context, key, value string) error {
	if key == c.failOn {
		return errors.New("simulated write error")
	}
	c.keys = append(c.keys, key)
	c.values = append(c.values, value)
	return nil
}

func TestPriorityWriter_PanicOnNilInner(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner writer")
		}
	}()
	sync.NewPriorityWriter(nil)
}

func TestPriorityWriter_FlushesInPriorityOrder(t *testing.T) {
	cap := &captureWriter{}
	pw := sync.NewPriorityWriter(cap)
	ctx := context.Background()

	_ = pw.WriteWithPriority(ctx, "LOW_KEY", "low", sync.PriorityLow)
	_ = pw.WriteWithPriority(ctx, "HIGH_KEY", "high", sync.PriorityHigh)
	_ = pw.WriteWithPriority(ctx, "NORMAL_KEY", "normal", sync.PriorityNormal)

	if err := pw.Flush(ctx); err != nil {
		t.Fatalf("unexpected flush error: %v", err)
	}

	want := []string{"HIGH_KEY", "NORMAL_KEY", "LOW_KEY"}
	for i, k := range want {
		if cap.keys[i] != k {
			t.Errorf("position %d: got %q, want %q", i, cap.keys[i], k)
		}
	}
}

func TestPriorityWriter_FlushClearsBuffer(t *testing.T) {
	cap := &captureWriter{}
	pw := sync.NewPriorityWriter(cap)
	ctx := context.Background()

	_ = pw.Write(ctx, "KEY", "val")
	_ = pw.Flush(ctx)

	if pw.Len() != 0 {
		t.Errorf("expected empty buffer after flush, got %d entries", pw.Len())
	}
}

func TestPriorityWriter_StopsOnInnerError(t *testing.T) {
	cap := &captureWriter{failOn: "BAD_KEY"}
	pw := sync.NewPriorityWriter(cap)
	ctx := context.Background()

	_ = pw.WriteWithPriority(ctx, "BAD_KEY", "v", sync.PriorityHigh)
	_ = pw.WriteWithPriority(ctx, "GOOD_KEY", "v", sync.PriorityNormal)

	err := pw.Flush(ctx)
	if err == nil {
		t.Fatal("expected error from inner writer")
	}
	if len(cap.keys) != 0 {
		t.Errorf("expected no successful writes before error, got %v", cap.keys)
	}
}

func TestPriorityWriter_RejectEmptyKey(t *testing.T) {
	cap := &captureWriter{}
	pw := sync.NewPriorityWriter(cap)
	ctx := context.Background()

	if err := pw.Write(ctx, "", "value"); err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestPriorityWriter_Reset_DiscardsEntries(t *testing.T) {
	cap := &captureWriter{}
	pw := sync.NewPriorityWriter(cap)
	ctx := context.Background()

	_ = pw.Write(ctx, "KEY", "val")
	pw.Reset()

	if pw.Len() != 0 {
		t.Errorf("expected 0 entries after reset, got %d", pw.Len())
	}
	_ = pw.Flush(ctx)
	if len(cap.keys) != 0 {
		t.Errorf("expected no writes after reset+flush, got %v", cap.keys)
	}
}
