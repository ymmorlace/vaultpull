package sync

import (
	"context"
	"fmt"
	"sync"
)

// RollbackEntry holds the previous state of a secret key before it was written.
type RollbackEntry struct {
	Key      string
	OldValue string
	HadValue bool
}

// RollbackWriter wraps an EnvWriter and records writes so they can be undone.
type RollbackWriter struct {
	mu      sync.Mutex
	inner   EnvWriter
	log     []RollbackEntry
	snapshot map[string]string
}

// NewRollbackWriter creates a RollbackWriter that wraps inner.
// snapshot is the pre-existing state of the env file (may be nil or empty).
func NewRollbackWriter(inner EnvWriter, snapshot map[string]string) *RollbackWriter {
	if inner == nil {
		panic("rollback: inner writer must not be nil")
	}
	prev := make(map[string]string, len(snapshot))
	for k, v := range snapshot {
		prev[k] = v
	}
	return &RollbackWriter{
		inner:    inner,
		snapshot: prev,
	}
}

// Write records the previous value for key and delegates to the inner writer.
func (r *RollbackWriter) Write(ctx context.Context, key, value string) error {
	r.mu.Lock()
	old, had := r.snapshot[key]
	r.log = append(r.log, RollbackEntry{Key: key, OldValue: old, HadValue: had})
	r.mu.Unlock()

	if err := r.inner.Write(ctx, key, value); err != nil {
		return fmt.Errorf("rollback writer: %w", err)
	}
	return nil
}

// Entries returns a copy of the rollback log in write order.
func (r *RollbackWriter) Entries() []RollbackEntry {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]RollbackEntry, len(r.log))
	copy(out, r.log)
	return out
}

// Reset clears the rollback log without affecting the inner writer.
func (r *RollbackWriter) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.log = r.log[:0]
}
