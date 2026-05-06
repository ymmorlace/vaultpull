package sync

import (
	"context"
	"fmt"
)

// EnvWriter is the interface used throughout the sync package.
type splitterWriter interface {
	Write(ctx context.Context, key, value string) error
}

// RouteFn decides which writer index should receive a given key.
// It returns -1 to broadcast to all writers.
type RouteFn func(key string) int

// Splitter fans out writes to multiple writers based on a routing function.
// If the RouteFn returns -1 the entry is broadcast to every writer.
type Splitter struct {
	writers []splitterWriter
	routeFn RouteFn
}

// NewSplitter creates a Splitter that routes writes using routeFn.
// At least one writer must be provided.
func NewSplitter(routeFn RouteFn, writers ...splitterWriter) *Splitter {
	if len(writers) == 0 {
		panic("splitter: at least one writer is required")
	}
	if routeFn == nil {
		panic("splitter: routeFn must not be nil")
	}
	return &Splitter{writers: writers, routeFn: routeFn}
}

// Write routes the key/value pair to the appropriate writer(s).
func (s *Splitter) Write(ctx context.Context, key, value string) error {
	idx := s.routeFn(key)
	if idx == -1 {
		for i, w := range s.writers {
			if err := w.Write(ctx, key, value); err != nil {
				return fmt.Errorf("splitter: writer %d: %w", i, err)
			}
		}
		return nil
	}
	if idx < 0 || idx >= len(s.writers) {
		return fmt.Errorf("splitter: route index %d out of range [0, %d)", idx, len(s.writers))
	}
	return s.writers[idx].Write(ctx, key, value)
}

// Len returns the number of backing writers.
func (s *Splitter) Len() int { return len(s.writers) }
