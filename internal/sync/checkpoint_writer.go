package sync

import "context"

// EnvWriter is the minimal interface expected by checkpoint-aware writers.
// It matches the interface used throughout the sync package.
type checkpointEnvWriter interface {
	Write(ctx context.Context, path, key, value string) error
}

// CheckpointWriter wraps an inner EnvWriter and skips entries that have
// already been recorded in the provided Checkpoint.
type CheckpointWriter struct {
	inner checkpointEnvWriter
	cp    *Checkpoint
}

// NewCheckpointWriter returns a writer that consults cp before delegating
// to inner. Entries already present in cp are silently skipped.
func NewCheckpointWriter(inner checkpointEnvWriter, cp *Checkpoint) *CheckpointWriter {
	if inner == nil {
		panic("checkpointwriter: inner writer must not be nil")
	}
	if cp == nil {
		panic("checkpointwriter: checkpoint must not be nil")
	}
	return &CheckpointWriter{inner: inner, cp: cp}
}

// Write skips the entry when it is already present in the checkpoint;
// otherwise it delegates to the inner writer and, on success, records the
// entry so that subsequent calls are skipped.
func (w *CheckpointWriter) Write(ctx context.Context, path, key, value string) error {
	if w.cp.Has(path, key) {
		return nil
	}
	if err := w.inner.Write(ctx, path, key, value); err != nil {
		return err
	}
	w.cp.Record(path, key)
	return nil
}
