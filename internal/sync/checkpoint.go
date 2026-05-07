package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// CheckpointEntry records the last-synced state of a single secret path.
type CheckpointEntry struct {
	Path      string    `json:"path"`
	Key       string    `json:"key"`
	SyncedAt  time.Time `json:"synced_at"`
}

// Checkpoint tracks which secrets have been successfully written so that
// incremental syncs can skip unchanged entries.
type Checkpoint struct {
	mu      sync.RWMutex
	entries map[string]CheckpointEntry // keyed by path+key
}

// NewCheckpoint returns an empty in-memory checkpoint.
func NewCheckpoint() *Checkpoint {
	return &Checkpoint{
		entries: make(map[string]CheckpointEntry),
	}
}

func checkpointKey(path, key string) string {
	return path + "\x00" + key
}

// Record marks the given path/key pair as synced at now.
func (c *Checkpoint) Record(path, key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[checkpointKey(path, key)] = CheckpointEntry{
		Path:     path,
		Key:      key,
		SyncedAt: time.Now().UTC(),
	}
}

// Has reports whether the path/key pair has been previously recorded.
func (c *Checkpoint) Has(path, key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.entries[checkpointKey(path, key)]
	return ok
}

// Reset clears all recorded entries.
func (c *Checkpoint) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]CheckpointEntry)
}

// Save persists the checkpoint to the given file path as JSON.
func (c *Checkpoint) Save(filePath string) error {
	c.mu.RLock()
	list := make([]CheckpointEntry, 0, len(c.entries))
	for _, e := range c.entries {
		list = append(list, e)
	}
	c.mu.RUnlock()

	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return fmt.Errorf("checkpoint marshal: %w", err)
	}
	return os.WriteFile(filePath, data, 0o600)
}

// LoadCheckpoint reads a previously saved checkpoint from disk.
func LoadCheckpoint(filePath string) (*Checkpoint, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("checkpoint read: %w", err)
	}
	var list []CheckpointEntry
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("checkpoint unmarshal: %w", err)
	}
	cp := NewCheckpoint()
	for _, e := range list {
		cp.entries[checkpointKey(e.Path, e.Key)] = e
	}
	return cp, nil
}
