package sync_test

import (
	"context"
	"errors"
	"testing"

	sync "github.com/yourusername/vaultpull/internal/sync"
)

func TestBufferedWriter_PanicOnNilInner(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner writer")
		}
	}()
	sync.NewBufferedWriter(nil, 4)
}

func TestBufferedWriter_PanicOnZeroCapacity(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero capacity")
		}
	}()
	sync.NewBufferedWriter(&captureWriter{}, 0)
}

func TestBufferedWriter_BuffersWithoutFlush(t *testing.T) {
	cw := &captureWriter{}
	bw := sync.NewBufferedWriter(cw, 4)
	ctx := context.Background()

	_ = bw.Write(ctx, "KEY1", "val1")
	_ = bw.Write(ctx, "KEY2", "val2")

	if cw.count() != 0 {
		t.Fatalf("expected 0 inner writes before flush, got %d", cw.count())
	}
	if bw.Len() != 2 {
		t.Fatalf("expected buffer length 2, got %d", bw.Len())
	}
}

func TestBufferedWriter_FlushWritesAll(t *testing.T) {
	cw := &captureWriter{}
	bw := sync.NewBufferedWriter(cw, 4)
	ctx := context.Background()

	_ = bw.Write(ctx, "A", "1")
	_ = bw.Write(ctx, "B", "2")
	if err := bw.Flush(ctx); err != nil {
		t.Fatalf("unexpected flush error: %v", err)
	}

	if cw.count() != 2 {
		t.Fatalf("expected 2 inner writes after flush, got %d", cw.count())
	}
	if bw.Len() != 0 {
		t.Fatalf("expected buffer empty after flush, got %d", bw.Len())
	}
}

func TestBufferedWriter_AutoFlushOnFull(t *testing.T) {
	cw := &captureWriter{}
	bw := sync.NewBufferedWriter(cw, 2)
	ctx := context.Background()

	_ = bw.Write(ctx, "X", "1")
	_ = bw.Write(ctx, "Y", "2")
	// buffer is full; next write triggers auto-flush
	_ = bw.Write(ctx, "Z", "3")

	if cw.count() != 2 {
		t.Fatalf("expected 2 flushed writes, got %d", cw.count())
	}
	if bw.Len() != 1 {
		t.Fatalf("expected 1 buffered entry after auto-flush, got %d", bw.Len())
	}
}

func TestBufferedWriter_PropagatesInnerError(t *testing.T) {
	sentinel := errors.New("write failed")
	ew := &errorWriter{err: sentinel}
	bw := sync.NewBufferedWriter(ew, 2)
	ctx := context.Background()

	_ = bw.Write(ctx, "A", "1")
	_ = bw.Write(ctx, "B", "2")
	// auto-flush on third write; inner returns error
	err := bw.Write(ctx, "C", "3")
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestBufferedWriter_FlushClearsOnSuccess(t *testing.T) {
	cw := &captureWriter{}
	bw := sync.NewBufferedWriter(cw, 10)
	ctx := context.Background()

	_ = bw.Write(ctx, "K", "v")
	_ = bw.Flush(ctx)
	_ = bw.Flush(ctx) // second flush should be a no-op

	if cw.count() != 1 {
		t.Fatalf("expected 1 write total, got %d", cw.count())
	}
}
