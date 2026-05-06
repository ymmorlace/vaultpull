package sync_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/yourusername/vaultpull/internal/sync"
)

func TestCompressedWriter_BuffersEntries(t *testing.T) {
	inner := &captureWriter{}
	cw := sync.NewCompressedWriter(inner)

	ctx := context.Background()
	_ = cw.Write(ctx, "FOO", "bar")
	_ = cw.Write(ctx, "BAZ", "qux")

	if cw.Len() != 2 {
		t.Fatalf("expected 2 buffered entries, got %d", cw.Len())
	}
}

func TestCompressedWriter_FlushWritesInner(t *testing.T) {
	inner := &captureWriter{}
	cw := sync.NewCompressedWriter(inner)

	ctx := context.Background()
	_ = cw.Write(ctx, "KEY", "value")
	_, err := cw.Flush(ctx)
	if err != nil {
		t.Fatalf("unexpected flush error: %v", err)
	}

	if len(inner.entries) != 1 || inner.entries[0].key != "KEY" {
		t.Fatalf("inner writer not called correctly: %+v", inner.entries)
	}
}

func TestCompressedWriter_FlushProducesValidGzip(t *testing.T) {
	inner := &captureWriter{}
	cw := sync.NewCompressedWriter(inner)

	ctx := context.Background()
	_ = cw.Write(ctx, "HELLO", "world")
	compressed, err := cw.Flush(ctx)
	if err != nil {
		t.Fatalf("flush error: %v", err)
	}

	gr, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatalf("gzip reader error: %v", err)
	}
	defer gr.Close()
	out, _ := io.ReadAll(gr)
	if !strings.Contains(string(out), "HELLO=world") {
		t.Fatalf("expected HELLO=world in decompressed output, got: %s", out)
	}
}

func TestCompressedWriter_Reset_ClearsEntries(t *testing.T) {
	inner := &captureWriter{}
	cw := sync.NewCompressedWriter(inner)

	ctx := context.Background()
	_ = cw.Write(ctx, "A", "1")
	cw.Reset()

	if cw.Len() != 0 {
		t.Fatalf("expected 0 entries after reset, got %d", cw.Len())
	}
}

func TestCompressedWriter_ContextCancelled(t *testing.T) {
	inner := &captureWriter{}
	cw := sync.NewCompressedWriter(inner)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := cw.Write(ctx, "X", "y")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestCompressedWriter_PanicOnNilInner(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nil inner writer")
		}
	}()
	sync.NewCompressedWriter(nil)
}

func TestCompressedWriter_FlushPropagatesInnerError(t *testing.T) {
	sentinel := errors.New("inner write failed")
	inner := &captureWriter{err: sentinel}
	cw := sync.NewCompressedWriter(inner)

	ctx := context.Background()
	_ = cw.Write(ctx, "ERR", "val")
	_, err := cw.Flush(ctx)
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}
