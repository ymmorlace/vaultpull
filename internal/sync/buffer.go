package sync

import (
	"context"
	"fmt"
	"sync"
)

// BufferedWriter accumulates writes up to a fixed capacity and flushes them
// to an inner EnvWriter in a single batch. If the buffer is full, Flush is
// called automatically before accepting the new entry.
type BufferedWriter struct {
	mu       sync.Mutex
	inner    EnvWriter
	entries  []envEntry
	capacity int
}

type envEntry struct {
	key   string
	value string
}

// NewBufferedWriter creates a BufferedWriter with the given capacity.
// Panics if inner is nil or capacity is less than 1.
func NewBufferedWriter(inner EnvWriter, capacity int) *BufferedWriter {
	if inner == nil {
		panic("bufferedwriter: inner writer must not be nil")
	}
	if capacity < 1 {
		panic("bufferedwriter: capacity must be at least 1")
	}
	return &BufferedWriter{
		inner:    inner,
		capacity: capacity,
		entries:  make([]envEntry, 0, capacity),
	}
}

// Write buffers the key/value pair. If the buffer is full it is flushed first.
func (b *BufferedWriter) Write(ctx context.Context, key, value string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.entries) >= b.capacity {
		if err := b.flush(ctx); err != nil {
			return fmt.Errorf("bufferedwriter: auto-flush: %w", err)
		}
	}
	b.entries = append(b.entries, envEntry{key: key, value: value})
	return nil
}

// Flush writes all buffered entries to the inner writer and clears the buffer.
func (b *BufferedWriter) Flush(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.flush(ctx)
}

// Len returns the number of currently buffered entries.
func (b *BufferedWriter) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.entries)
}

func (b *BufferedWriter) flush(ctx context.Context) error {
	for _, e := range b.entries {
		if err := b.inner.Write(ctx, e.key, e.value); err != nil {
			return err
		}
	}
	b.entries = b.entries[:0]
	return nil
}
