package sync

import (
	"compress/gzip"
	"bytes"
	"context"
	"fmt"
	"io"
)

// CompressedWriter wraps an EnvWriter and compresses snapshot data
// before passing it downstream. It is useful when persisting large
// secret sets to disk or a remote store.
type CompressedWriter struct {
	inner  EnvWriter
	buf    bytes.Buffer
	entries []envEntry
}

type envEntry struct {
	key   string
	value string
}

// NewCompressedWriter returns a CompressedWriter that delegates to inner
// after accumulating entries. Call Flush to compress and write all entries.
func NewCompressedWriter(inner EnvWriter) *CompressedWriter {
	if inner == nil {
		panic("compressor: inner writer must not be nil")
	}
	return &CompressedWriter{inner: inner}
}

// Write records the key/value pair for later flushing.
func (c *CompressedWriter) Write(ctx context.Context, key, value string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	c.entries = append(c.entries, envEntry{key: key, value: value})
	return nil
}

// Flush compresses the accumulated entries and writes them via the inner writer.
func (c *CompressedWriter) Flush(ctx context.Context) ([]byte, error) {
	c.buf.Reset()
	gw := gzip.NewWriter(&c.buf)
	for _, e := range c.entries {
		line := fmt.Sprintf("%s=%s\n", e.key, e.value)
		if _, err := io.WriteString(gw, line); err != nil {
			return nil, fmt.Errorf("compressor: gzip write: %w", err)
		}
		if err := c.inner.Write(ctx, e.key, e.value); err != nil {
			return nil, err
		}
	}
	if err := gw.Close(); err != nil {
		return nil, fmt.Errorf("compressor: gzip close: %w", err)
	}
	return c.buf.Bytes(), nil
}

// Reset clears buffered entries without writing them.
func (c *CompressedWriter) Reset() {
	c.entries = c.entries[:0]
	c.buf.Reset()
}

// Len returns the number of buffered entries.
func (c *CompressedWriter) Len() int {
	return len(c.entries)
}
