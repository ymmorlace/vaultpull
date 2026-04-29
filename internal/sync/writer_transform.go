package sync

import (
	"fmt"
	"os"
	"strings"
)

// TransformingWriter wraps an EnvWriter and applies a ValueTransformer to each
// value before delegating to the underlying writer.
type TransformingWriter struct {
	inner       EnvWriter
	transformer ValueTransformer
}

// EnvWriter is the interface for writing key-value env entries.
type EnvWriter interface {
	Write(key, value string) error
	Close() error
}

// NewTransformingWriter creates a TransformingWriter.
func NewTransformingWriter(inner EnvWriter, transformer ValueTransformer) *TransformingWriter {
	return &TransformingWriter{inner: inner, transformer: transformer}
}

// Write transforms the value then delegates to the inner writer.
func (tw *TransformingWriter) Write(key, value string) error {
	transformed, err := tw.transformer.Transform(key, value)
	if err != nil {
		return fmt.Errorf("transform failed for %q: %w", key, err)
	}
	return tw.inner.Write(key, transformed)
}

// Close closes the underlying writer.
func (tw *TransformingWriter) Close() error {
	return tw.inner.Close()
}

// fileEnvWriterAdapter adapts FileEnvWriter to satisfy EnvWriter.
type fileEnvWriterAdapter struct {
	path string
	f    *os.File
}

// NewFileEnvWriterAdapter opens a file and returns an EnvWriter.
func NewFileEnvWriterAdapter(path string) (EnvWriter, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}
	return &fileEnvWriterAdapter{path: path, f: f}, nil
}

// Write writes KEY=value to the file.
func (a *fileEnvWriterAdapter) Write(key, value string) error {
	_, err := fmt.Fprintf(a.f, "%s=%s\n", strings.ToUpper(key), value)
	return err
}

// Close closes the underlying file.
func (a *fileEnvWriterAdapter) Close() error {
	return a.f.Close()
}
