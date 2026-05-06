package sync

import (
	"context"
	"fmt"
	"io"
	"os"
)

// DryRunWriter is an EnvWriter that records what would be written
// without performing any actual I/O. It reports each entry to a
// human-readable log writer instead.
type DryRunWriter struct {
	out    io.Writer
	entries []DryRunEntry
}

// DryRunEntry captures a single would-be write operation.
type DryRunEntry struct {
	Key   string
	Value string
}

// NewDryRunWriter returns a DryRunWriter that prints to out.
// If out is nil, os.Stdout is used.
func NewDryRunWriter(out io.Writer) *DryRunWriter {
	if out == nil {
		out = os.Stdout
	}
	return &DryRunWriter{out: out}
}

// Write records the entry and prints a dry-run notice instead of
// persisting anything.
func (d *DryRunWriter) Write(_ context.Context, key, value string) error {
	d.entries = append(d.entries, DryRunEntry{Key: key, Value: value})
	_, err := fmt.Fprintf(d.out, "[dry-run] would write %s=%q\n", key, value)
	return err
}

// Entries returns all entries that were presented to Write.
func (d *DryRunWriter) Entries() []DryRunEntry {
	result := make([]DryRunEntry, len(d.entries))
	copy(result, d.entries)
	return result
}

// Reset clears the recorded entries so the writer can be reused.
func (d *DryRunWriter) Reset() {
	d.entries = d.entries[:0]
}
