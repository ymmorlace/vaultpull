package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// SnapshotEntry records a single secret written during a sync run.
type SnapshotEntry struct {
	Path  string `json:"path"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Snapshot captures the full state of secrets written in a sync run.
type Snapshot struct {
	Timestamp time.Time       `json:"timestamp"`
	Entries   []SnapshotEntry `json:"entries"`
}

// SnapshotWriter wraps an EnvWriter and records every written secret
// into an in-memory Snapshot that can be persisted to disk.
type SnapshotWriter struct {
	inner    EnvWriter
	snapshot *Snapshot
}

// NewSnapshotWriter returns a SnapshotWriter that delegates writes to inner
// and accumulates entries into a new Snapshot.
func NewSnapshotWriter(inner EnvWriter) *SnapshotWriter {
	return &SnapshotWriter{
		inner: inner,
		snapshot: &Snapshot{
			Timestamp: time.Now().UTC(),
			Entries:   []SnapshotEntry{},
		},
	}
}

// Write delegates to the inner writer and records the entry on success.
func (s *SnapshotWriter) Write(path, key, value string) error {
	if err := s.inner.Write(path, key, value); err != nil {
		return err
	}
	s.snapshot.Entries = append(s.snapshot.Entries, SnapshotEntry{
		Path:  path,
		Key:   key,
		Value: value,
	})
	return nil
}

// Snapshot returns the accumulated snapshot.
func (s *SnapshotWriter) Snapshot() *Snapshot {
	return s.snapshot
}

// SaveSnapshot writes the snapshot as JSON to the given file path.
func SaveSnapshot(snap *Snapshot, filePath string) error {
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("snapshot: open %s: %w", filePath, err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(snap); err != nil {
		return fmt.Errorf("snapshot: encode: %w", err)
	}
	return nil
}
