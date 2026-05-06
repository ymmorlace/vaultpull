package sync

import (
	"context"
	"errors"
	"testing"
)

func TestDedupWriter_PanicOnNilInner(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner writer")
		}
	}()
	NewDedupWriter(nil)
}

func TestDedupWriter_FirstWriteSucceeds(t *testing.T) {
	cap := &captureWriter{}
	w := NewDedupWriter(cap)

	if err := w.Write(context.Background(), "secret/app", "API_KEY", "abc"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cap.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(cap.entries))
	}
}

func TestDedupWriter_DuplicateKeyReturnsError(t *testing.T) {
	cap := &captureWriter{}
	w := NewDedupWriter(cap)

	_ = w.Write(context.Background(), "secret/app", "API_KEY", "first")
	err := w.Write(context.Background(), "secret/app", "API_KEY", "second")

	if err == nil {
		t.Fatal("expected error for duplicate key, got nil")
	}
	// Inner writer should only have received the first write.
	if len(cap.entries) != 1 {
		t.Fatalf("expected 1 entry in inner writer, got %d", len(cap.entries))
	}
}

func TestDedupWriter_InnerErrorDoesNotRecordKey(t *testing.T) {
	sentinel := errors.New("write failed")
	w := NewDedupWriter(&errorWriter{err: sentinel})

	_ = w.Write(context.Background(), "secret/app", "DB_PASS", "x")
	// Second attempt should not be treated as a duplicate.
	err := w.Write(context.Background(), "secret/app", "DB_PASS", "x")

	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got: %v", err)
	}
}

func TestDedupWriter_Reset_AllowsReuse(t *testing.T) {
	cap := &captureWriter{}
	w := NewDedupWriter(cap)

	_ = w.Write(context.Background(), "secret/app", "TOKEN", "v1")
	w.Reset()

	if err := w.Write(context.Background(), "secret/app", "TOKEN", "v2"); err != nil {
		t.Fatalf("expected no error after reset, got: %v", err)
	}
	if len(cap.entries) != 2 {
		t.Fatalf("expected 2 entries after reset, got %d", len(cap.entries))
	}
}

func TestDedupWriter_DifferentPathsSameKey(t *testing.T) {
	cap := &captureWriter{}
	w := NewDedupWriter(cap)

	_ = w.Write(context.Background(), "secret/app", "PORT", "8080")
	err := w.Write(context.Background(), "secret/db", "PORT", "5432")

	if err == nil {
		t.Fatal("expected conflict error for same key from different paths")
	}
}
