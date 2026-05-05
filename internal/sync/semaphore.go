package sync

import (
	"context"
	"fmt"
)

// Semaphore limits the number of concurrent secret writes.
type Semaphore struct {
	slots chan struct{}
}

// NewSemaphore creates a Semaphore with the given concurrency limit.
// It panics if limit is less than 1.
func NewSemaphore(limit int) *Semaphore {
	if limit < 1 {
		panic(fmt.Sprintf("semaphore: limit must be >= 1, got %d", limit))
	}
	return &Semaphore{
		slots: make(chan struct{}, limit),
	}
}

// Acquire blocks until a slot is available or the context is cancelled.
func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case s.slots <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release frees one slot. It is a no-op if the semaphore is already empty.
func (s *Semaphore) Release() {
	select {
	case <-s.slots:
	default:
	}
}

// Available returns the number of slots currently free.
func (s *Semaphore) Available() int {
	return cap(s.slots) - len(s.slots)
}

// NewSemaphoreWriter wraps an EnvWriter and enforces a concurrency limit
// across concurrent calls to Write.
func NewSemaphoreWriter(inner EnvWriter, limit int) EnvWriter {
	sem := NewSemaphore(limit)
	return &semaphoreWriter{inner: inner, sem: sem}
}

type semaphoreWriter struct {
	inner EnvWriter
	sem   *Semaphore
}

func (w *semaphoreWriter) Write(ctx context.Context, key, value string) error {
	if err := w.sem.Acquire(ctx); err != nil {
		return fmt.Errorf("semaphore acquire: %w", err)
	}
	defer w.sem.Release()
	return w.inner.Write(ctx, key, value)
}
