package sync

import (
	"context"
	"sort"
)

// SortOrder defines the ordering direction for keys.
type SortOrder int

const (
	// SortAscending sorts keys from A to Z.
	SortAscending SortOrder = iota
	// SortDescending sorts keys from Z to A.
	SortDescending
)

// SortedWriter buffers all written entries and flushes them to the inner
// EnvWriter in sorted key order when Flush is called.
type SortedWriter struct {
	inner   EnvWriter
	order   SortOrder
	entries []envEntry
}

type envEntry struct {
	key   string
	value string
}

// NewSortedWriter returns a SortedWriter that delegates to inner after sorting.
// Panics if inner is nil.
func NewSortedWriter(inner EnvWriter, order SortOrder) *SortedWriter {
	if inner == nil {
		panic("sorter: inner writer must not be nil")
	}
	return &SortedWriter{inner: inner, order: order}
}

// Write buffers the key/value pair for later sorted flushing.
func (s *SortedWriter) Write(_ context.Context, key, value string) error {
	s.entries = append(s.entries, envEntry{key: key, value: value})
	return nil
}

// Flush sorts buffered entries and writes them to the inner writer in order.
// Any write error stops flushing and returns immediately.
func (s *SortedWriter) Flush(ctx context.Context) error {
	sorted := make([]envEntry, len(s.entries))
	copy(sorted, s.entries)

	sort.SliceStable(sorted, func(i, j int) bool {
		if s.order == SortDescending {
			return sorted[i].key > sorted[j].key
		}
		return sorted[i].key < sorted[j].key
	})

	for _, e := range sorted {
		if err := s.inner.Write(ctx, e.key, e.value); err != nil {
			return err
		}
	}
	return nil
}

// Reset discards all buffered entries without flushing.
func (s *SortedWriter) Reset() {
	s.entries = s.entries[:0]
}

// Entries returns a copy of the currently buffered entries in insertion order.
func (s *SortedWriter) Entries() []envEntry {
	out := make([]envEntry, len(s.entries))
	copy(out, s.entries)
	return out
}
