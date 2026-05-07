package sync

import (
	"context"
	"errors"
	"testing"
)

func TestSortedWriter_PanicOnNilInner(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner writer")
		}
	}()
	NewSortedWriter(nil, SortAscending)
}

func TestSortedWriter_FlushAscending(t *testing.T) {
	cap := &captureWriter{}
	w := NewSortedWriter(cap, SortAscending)
	ctx := context.Background()

	_ = w.Write(ctx, "ZEBRA", "z")
	_ = w.Write(ctx, "APPLE", "a")
	_ = w.Write(ctx, "MANGO", "m")

	if err := w.Flush(ctx); err != nil {
		t.Fatalf("unexpected flush error: %v", err)
	}

	want := []string{"APPLE", "MANGO", "ZEBRA"}
	if len(cap.keys) != len(want) {
		t.Fatalf("got %d entries, want %d", len(cap.keys), len(want))
	}
	for i, k := range want {
		if cap.keys[i] != k {
			t.Errorf("position %d: got %q, want %q", i, cap.keys[i], k)
		}
	}
}

func TestSortedWriter_FlushDescending(t *testing.T) {
	cap := &captureWriter{}
	w := NewSortedWriter(cap, SortDescending)
	ctx := context.Background()

	_ = w.Write(ctx, "APPLE", "a")
	_ = w.Write(ctx, "ZEBRA", "z")
	_ = w.Write(ctx, "MANGO", "m")

	if err := w.Flush(ctx); err != nil {
		t.Fatalf("unexpected flush error: %v", err)
	}

	want := []string{"ZEBRA", "MANGO", "APPLE"}
	for i, k := range want {
		if cap.keys[i] != k {
			t.Errorf("position %d: got %q, want %q", i, cap.keys[i], k)
		}
	}
}

func TestSortedWriter_StopsOnInnerError(t *testing.T) {
	sentinel := errors.New("write failed")
	ew := &errorAfterWriter{failAfter: 1, err: sentinel}
	w := NewSortedWriter(ew, SortAscending)
	ctx := context.Background()

	_ = w.Write(ctx, "A", "1")
	_ = w.Write(ctx, "B", "2")
	_ = w.Write(ctx, "C", "3")

	if err := w.Flush(ctx); !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if ew.count != 2 {
		t.Errorf("expected 2 writes before error, got %d", ew.count)
	}
}

func TestSortedWriter_Reset_ClearsEntries(t *testing.T) {
	cap := &captureWriter{}
	w := NewSortedWriter(cap, SortAscending)
	ctx := context.Background()

	_ = w.Write(ctx, "KEY", "val")
	w.Reset()

	_ = w.Flush(ctx)
	if len(cap.keys) != 0 {
		t.Errorf("expected no entries after reset, got %d", len(cap.keys))
	}
}

func TestSortedWriter_EntriesReturnsCopy(t *testing.T) {
	cap := &captureWriter{}
	w := NewSortedWriter(cap, SortAscending)
	ctx := context.Background()

	_ = w.Write(ctx, "FOO", "bar")
	e := w.Entries()
	e[0].key = "MUTATED"

	if w.Entries()[0].key != "FOO" {
		t.Error("Entries should return a copy, not a reference")
	}
}

// errorAfterWriter fails after a given number of successful writes.
type errorAfterWriter struct {
	failAfter int
	err       error
	count     int
}

func (e *errorAfterWriter) Write(_ context.Context, _, _ string) error {
	e.count++
	if e.count > e.failAfter {
		return e.err
	}
	return nil
}
