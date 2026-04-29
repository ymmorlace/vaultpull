package sync

import (
	"sync"
	"time"
)

// SyncMetrics tracks statistics for a sync run.
type SyncMetrics struct {
	mu        sync.Mutex
	Written   int
	Skipped   int
	Errors    int
	StartedAt time.Time
	FinishedAt time.Time
}

// NewSyncMetrics creates a new SyncMetrics instance with start time set.
func NewSyncMetrics() *SyncMetrics {
	return &SyncMetrics{
		StartedAt: time.Now(),
	}
}

// RecordWritten increments the written counter.
func (m *SyncMetrics) RecordWritten() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Written++
}

// RecordSkipped increments the skipped counter.
func (m *SyncMetrics) RecordSkipped() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Skipped++
}

// RecordError increments the error counter.
func (m *SyncMetrics) RecordError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Errors++
}

// Finish marks the end time of the sync run.
func (m *SyncMetrics) Finish() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.FinishedAt = time.Now()
}

// Duration returns the elapsed time between start and finish.
// If Finish has not been called, it returns the duration since start.
func (m *SyncMetrics) Duration() time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.FinishedAt.IsZero() {
		return time.Since(m.StartedAt)
	}
	return m.FinishedAt.Sub(m.StartedAt)
}

// Summary returns a snapshot of the current metrics.
func (m *SyncMetrics) Summary() MetricsSummary {
	m.mu.Lock()
	defer m.mu.Unlock()
	return MetricsSummary{
		Written:  m.Written,
		Skipped:  m.Skipped,
		Errors:   m.Errors,
		Duration: m.FinishedAt.Sub(m.StartedAt),
	}
}

// MetricsSummary is a value snapshot of SyncMetrics.
type MetricsSummary struct {
	Written  int
	Skipped  int
	Errors   int
	Duration time.Duration
}
