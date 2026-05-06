package sync

import (
	"context"
	"fmt"
	"sort"
)

// Priority levels for secret writes.
const (
	PriorityHigh   = 0
	PriorityNormal = 1
	PriorityLow    = 2
)

// priorityEntry holds a secret write job with its assigned priority.
type priorityEntry struct {
	key      string
	value    string
	priority int
}

// PriorityWriter buffers writes, orders them by priority, then flushes
// to an inner EnvWriter in priority order (lowest number = highest priority).
type PriorityWriter struct {
	inner   EnvWriter
	entries []priorityEntry
}

// NewPriorityWriter creates a PriorityWriter wrapping inner.
// Panics if inner is nil.
func NewPriorityWriter(inner EnvWriter) *PriorityWriter {
	if inner == nil {
		panic("prioritywriter: inner writer must not be nil")
	}
	return &PriorityWriter{inner: inner}
}

// WriteWithPriority enqueues a key/value pair at the given priority level.
func (p *PriorityWriter) WriteWithPriority(_ context.Context, key, value string, priority int) error {
	if key == "" {
		return fmt.Errorf("prioritywriter: key must not be empty")
	}
	p.entries = append(p.entries, priorityEntry{key: key, value: value, priority: priority})
	return nil
}

// Write enqueues a key/value pair at PriorityNormal.
func (p *PriorityWriter) Write(ctx context.Context, key, value string) error {
	return p.WriteWithPriority(ctx, key, value, PriorityNormal)
}

// Flush sorts buffered entries by priority and writes them to the inner writer.
// Stops and returns on the first error. Clears the buffer on success.
func (p *PriorityWriter) Flush(ctx context.Context) error {
	sort.SliceStable(p.entries, func(i, j int) bool {
		return p.entries[i].priority < p.entries[j].priority
	})
	for _, e := range p.entries {
		if err := p.inner.Write(ctx, e.key, e.value); err != nil {
			return fmt.Errorf("prioritywriter: flush %q: %w", e.key, err)
		}
	}
	p.entries = p.entries[:0]
	return nil
}

// Reset discards all buffered entries without writing them.
func (p *PriorityWriter) Reset() {
	p.entries = p.entries[:0]
}

// Len returns the number of buffered entries.
func (p *PriorityWriter) Len() int {
	return len(p.entries)
}
