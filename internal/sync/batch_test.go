package sync_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/your-org/vaultpull/internal/sync"
)

type recordingWriter struct {
	written []sync.EnvEntry
	failOn  string
}

func (r *recordingWriter) Write(_ context.Context, key, value string) error {
	if r.failOn != "" && key == r.failOn {
		return fmt.Errorf("simulated write error for %s", key)
	}
	r.written = append(r.written, sync.EnvEntry{Key: key, Value: value})
	return nil
}

func TestBatchWriter_PanicOnNilInner(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner writer")
		}
	}()
	sync.NewBatchWriter(nil)
}

func TestBatchWriter_WriteAll_Success(t *testing.T) {
	rw := &recordingWriter{}
	bw := sync.NewBatchWriter(rw)

	entries := []sync.EnvEntry{
		{Key: "FOO", Value: "bar"},
		{Key: "BAZ", Value: "qux"},
	}

	if err := bw.WriteAll(context.Background(), entries); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := len(rw.written); got != 2 {
		t.Fatalf("expected 2 written entries, got %d", got)
	}
}

func TestBatchWriter_WriteAll_StopsOnError(t *testing.T) {
	rw := &recordingWriter{failOn: "BAZ"}
	bw := sync.NewBatchWriter(rw)

	entries := []sync.EnvEntry{
		{Key: "FOO", Value: "1"},
		{Key: "BAZ", Value: "2"},
		{Key: "QUX", Value: "3"},
	}

	err := bw.WriteAll(context.Background(), entries)
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if !errors.Is(err, err) {
		t.Fatalf("unexpected error type: %v", err)
	}
	// Only FOO should have been written before the failure.
	if got := len(rw.written); got != 1 {
		t.Fatalf("expected 1 written entry before failure, got %d", got)
	}
}

func TestBatchWriter_WriteAll_CancelledContext(t *testing.T) {
	rw := &recordingWriter{}
	bw := sync.NewBatchWriter(rw)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	entries := []sync.EnvEntry{
		{Key: "FOO", Value: "bar"},
	}

	err := bw.WriteAll(ctx, entries)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}

func TestBatchWriter_Len(t *testing.T) {
	bw := sync.NewBatchWriter(&recordingWriter{})
	entries := []sync.EnvEntry{{Key: "A"}, {Key: "B"}, {Key: "C"}}
	if got := bw.Len(entries); got != 3 {
		t.Fatalf("expected Len 3, got %d", got)
	}
}
