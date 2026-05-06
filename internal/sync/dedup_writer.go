package sync

import (
	"context"
	"fmt"
)

// DedupWriter wraps an EnvWriter and rejects duplicate keys within a single
// sync run. It delegates to a Deduplicator for tracking and delegates writes
// to the inner writer only when the key is seen for the first time.
type DedupWriter struct {
	inner EnvWriter
	dedup *Deduplicator
}

// NewDedupWriter returns a DedupWriter that guards inner against duplicate
// key writes. Panics if inner is nil.
func NewDedupWriter(inner EnvWriter) *DedupWriter {
	if inner == nil {
		panic("dedup_writer: inner writer must not be nil")
	}
	return &DedupWriter{
		inner: inner,
		dedup: NewDeduplicator(),
	}
}

// Write checks whether key has already been written in this run. If so, it
// returns an error describing the conflict without calling the inner writer.
// Otherwise it records the key and delegates to the inner writer.
func (d *DedupWriter) Write(ctx context.Context, path, key, value string) error {
	if err := d.dedup.Check(path, key); err != nil {
		return fmt.Errorf("dedup_writer: %w", err)
	}
	if err := d.inner.Write(ctx, path, key, value); err != nil {
		// Do not record the key on inner failure so a retry can succeed.
		d.dedup.Remove(path, key)
		return err
	}
	return nil
}

// Reset clears all recorded keys so the writer can be reused for a new run.
func (d *DedupWriter) Reset() {
	d.dedup.Reset()
}
