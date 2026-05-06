package sync_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/example/vaultpull/internal/sync"
)

func TestDispatcher_PanicOnZeroWorkers(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero workers")
		}
	}()
	sync.NewDispatcher(&captureWriter{}, 0)
}

func TestDispatcher_WritesAllJobs(t *testing.T) {
	cw := &captureWriter{}
	d := sync.NewDispatcher(cw, 3)
	jobs := []sync.Job{
		{Path: "app/db", Key: "DB_HOST", Value: "localhost"},
		{Path: "app/db", Key: "DB_PORT", Value: "5432"},
		{Path: "app/cache", Key: "REDIS_URL", Value: "redis://localhost"},
	}
	results := d.Dispatch(context.Background(), jobs)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error for %s: %v", r.Key, r.Err)
		}
	}
	if int(cw.count.Load()) != 3 {
		t.Errorf("expected 3 writes, got %d", cw.count.Load())
	}
}

func TestDispatcher_PropagatesWriteError(t *testing.T) {
	failErr := errors.New("write failed")
	fw := &failWriter{err: failErr}
	d := sync.NewDispatcher(fw, 1)
	jobs := []sync.Job{{Path: "p", Key: "KEY", Value: "val"}}
	results := d.Dispatch(context.Background(), jobs)
	if results[0].Err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(results[0].Err, failErr) {
		t.Errorf("expected wrapped failErr, got %v", results[0].Err)
	}
}

func TestDispatcher_ContextCancelled(t *testing.T) {
	cw := &captureWriter{}
	d := sync.NewDispatcher(cw, 2)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	jobs := []sync.Job{
		{Path: "p", Key: "A", Value: "1"},
		{Path: "p", Key: "B", Value: "2"},
	}
	results := d.Dispatch(ctx, jobs)
	for _, r := range results {
		if r.Err == nil {
			t.Errorf("expected context error for key %s", r.Key)
		}
	}
}

func TestDispatcher_EmptyJobs(t *testing.T) {
	cw := &captureWriter{}
	d := sync.NewDispatcher(cw, 2)
	results := d.Dispatch(context.Background(), nil)
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

// captureWriter counts successful writes.
type captureWriter struct {
	count atomic.Int64
}

func (c *captureWriter) Write(_ context.Context, _, _ string) error {
	c.count.Add(1)
	return nil
}

// failWriter always returns an error.
type failWriter struct{ err error }

func (f *failWriter) Write(_ context.Context, _, _ string) error { return f.err }
