package sync

import (
	"context"
	"fmt"
)

// EnvEntry represents a single key/value secret pair to be written.
type EnvEntry struct {
	Key   string
	Value string
	Path  string
}

// BatchWriter writes multiple EnvEntry values to an EnvWriter in a single
// logical operation, stopping on the first error.
type BatchWriter struct {
	inner EnvWriter
}

// NewBatchWriter wraps an EnvWriter so that slices of EnvEntry can be written
// atomically (all-or-nothing with respect to error propagation).
func NewBatchWriter(inner EnvWriter) *BatchWriter {
	if inner == nil {
		panic("batch: inner writer must not be nil")
	}
	return &BatchWriter{inner: inner}
}

// WriteAll writes every entry in the slice to the underlying EnvWriter.
// It returns the index and error of the first failing write, or nil when all
// entries succeed.
func (b *BatchWriter) WriteAll(ctx context.Context, entries []EnvEntry) error {
	for i, e := range entries {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("batch: context cancelled before entry %d (%s): %w", i, e.Key, err)
		}
		if err := b.inner.Write(ctx, e.Key, e.Value); err != nil {
			return fmt.Errorf("batch: failed to write entry %d (%s): %w", i, e.Key, err)
		}
	}
	return nil
}

// Len returns the number of entries that would be written.
func (b *BatchWriter) Len(entries []EnvEntry) int {
	return len(entries)
}
