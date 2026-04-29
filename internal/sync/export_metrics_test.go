package sync

// NewMetricsWriterExported exposes NewMetricsWriter for black-box tests.
func NewMetricsWriterExported(inner EnvWriter, m *SyncMetrics) *MetricsWriter {
	return NewMetricsWriter(inner, m)
}

// NewSyncMetricsExported exposes NewSyncMetrics for black-box tests.
func NewSyncMetricsExported() *SyncMetrics {
	return NewSyncMetrics()
}
