package sync

import (
	"context"
	"fmt"
	"strings"
)

// EnvelopeWriter wraps each written key with a configurable prefix and suffix,
// producing entries like: PREFIX__KEY__SUFFIX=value.
// This is useful when secrets need to be namespaced within a single .env file.
type EnvelopeWriter struct {
	inner  EnvWriter
	prefix string
	suffix string
	sep    string
}

// EnvWriter is the interface expected by EnvelopeWriter.
type EnvWriter interface {
	Write(ctx context.Context, key, value string) error
}

// NewEnvelopeWriter creates an EnvelopeWriter that wraps keys with the given
// prefix and suffix using sep as the separator. Panics if inner is nil or sep
// is empty.
func NewEnvelopeWriter(inner EnvWriter, prefix, suffix, sep string) *EnvelopeWriter {
	if inner == nil {
		panic("envelope: inner writer must not be nil")
	}
	if sep == "" {
		panic("envelope: separator must not be empty")
	}
	return &EnvelopeWriter{
		inner:  inner,
		prefix: prefix,
		suffix: suffix,
		sep:    sep,
	}
}

// Write wraps key with the configured prefix and suffix before delegating to
// the inner writer.
func (e *EnvelopeWriter) Write(ctx context.Context, key, value string) error {
	wrapped := e.buildKey(key)
	return e.inner.Write(ctx, wrapped, value)
}

// buildKey assembles the final key from prefix, original key, and suffix,
// omitting empty segments.
func (e *EnvelopeWriter) buildKey(key string) string {
	parts := make([]string, 0, 3)
	if e.prefix != "" {
		parts = append(parts, e.prefix)
	}
	parts = append(parts, key)
	if e.suffix != "" {
		parts = append(parts, e.suffix)
	}
	return strings.Join(parts, e.sep)
}

// Describe returns a human-readable description of the envelope configuration.
func (e *EnvelopeWriter) Describe() string {
	return fmt.Sprintf("EnvelopeWriter(prefix=%q, suffix=%q, sep=%q)", e.prefix, e.suffix, e.sep)
}
