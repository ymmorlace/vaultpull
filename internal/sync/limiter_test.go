package sync_test

import (
	"context"
	"errors"
	"testing"

	internalsync "github.com/yourusername/vaultpull/internal/sync"
)

func TestLimitWriter_PanicOnNilInner(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner writer")
		}
	}()
	internalsync.NewLimitWriter(nil, 5)
}

func TestLimitWriter_PanicOnZeroLimit(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero limit")
		}
	}()
	internalsync.NewLimitWriter(&captureWriter{}, 0)
}

func TestLimitWriter_PanicOnNegativeLimit(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for negative limit")
		}
	}()
	internalsync.NewLimitWriter(&captureWriter{}, -3)
}

func TestLimitWriter_AllowsUpToLimit(t *testing.T) {
	cw := &captureWriter{}
	lw := internalsync.NewLimitWriter(cw, 3)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		if err := lw.Write(ctx, "KEY", "val"); err != nil {
			t.Fatalf("write %d: unexpected error: %v", i, err)
		}
	}
	if lw.Count() != 3 {
		t.Fatalf("expected count 3, got %d", lw.Count())
	}
}

func TestLimitWriter_BlocksAfterLimit(t *testing.T) {
	cw := &captureWriter{}
	lw := internalsync.NewLimitWriter(cw, 2)
	ctx := context.Background()

	_ = lw.Write(ctx, "A", "1")
	_ = lw.Write(ctx, "B", "2")

	err := lw.Write(ctx, "C", "3")
	if !errors.Is(err, internalsync.ErrLimitExceeded) {
		t.Fatalf("expected ErrLimitExceeded, got %v", err)
	}
	if len(cw.entries) != 2 {
		t.Fatalf("inner writer should have received exactly 2 entries, got %d", len(cw.entries))
	}
}

func TestLimitWriter_InnerErrorDoesNotIncrementCount(t *testing.T) {
	sentinel := errors.New("inner failure")
	ew := &errorWriter{err: sentinel}
	lw := internalsync.NewLimitWriter(ew, 5)
	ctx := context.Background()

	if err := lw.Write(ctx, "K", "v"); !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if lw.Count() != 0 {
		t.Fatalf("count should remain 0 after inner error, got %d", lw.Count())
	}
}

func TestLimitWriter_Reset_AllowsReuse(t *testing.T) {
	cw := &captureWriter{}
	lw := internalsync.NewLimitWriter(cw, 1)
	ctx := context.Background()

	_ = lw.Write(ctx, "X", "1")
	if err := lw.Write(ctx, "Y", "2"); !errors.Is(err, internalsync.ErrLimitExceeded) {
		t.Fatal("expected limit exceeded before reset")
	}

	lw.Reset()
	if lw.Count() != 0 {
		t.Fatalf("count should be 0 after reset, got %d", lw.Count())
	}
	if err := lw.Write(ctx, "Z", "3"); err != nil {
		t.Fatalf("unexpected error after reset: %v", err)
	}
}
