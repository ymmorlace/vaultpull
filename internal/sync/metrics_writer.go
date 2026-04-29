package sync

// MetricsWriter wraps an EnvWriter and records write outcomes into SyncMetrics.
type MetricsWriter struct {
	inner   EnvWriter
	metrics *SyncMetrics
}

// EnvWriter is the interface satisfied by writers used in the sync pipeline.
type EnvWriter interface {
	Write(key, value string) error
}

// NewMetricsWriter creates a MetricsWriter that decorates inner with metric tracking.
func NewMetricsWriter(inner EnvWriter, metrics *SyncMetrics) *MetricsWriter {
	return &MetricsWriter{
		inner:   inner,
		metrics: metrics,
	}
}

// Write delegates to the inner writer and records the outcome.
func (w *MetricsWriter) Write(key, value string) error {
	err := w.inner.Write(key, value)
	if err != nil {
		w.metrics.RecordError()
		return err
	}
	w.metrics.RecordWritten()
	return nil
}
