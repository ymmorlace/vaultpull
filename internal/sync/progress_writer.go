package sync

import "context"

// ProgressWriter wraps an EnvWriter and reports progress for each write.
type ProgressWriter struct {
	inner    EnvWriter
	progress *ProgressReporter
}

// NewProgressWriter creates a ProgressWriter decorating inner.
func NewProgressWriter(inner EnvWriter, progress *ProgressReporter) *ProgressWriter {
	if inner == nil {
		panic("progress writer: inner must not be nil")
	}
	if progress == nil {
		panic("progress writer: progress reporter must not be nil")
	}
	return &ProgressWriter{inner: inner, progress: progress}
}

// Write delegates to the inner writer and records the outcome.
func (w *ProgressWriter) Write(ctx context.Context, key, value string) error {
	err := w.inner.Write(ctx, key, value)
	if err != nil {
		w.progress.RecordError(key, err)
		return err
	}
	w.progress.RecordWritten(key)
	return nil
}
