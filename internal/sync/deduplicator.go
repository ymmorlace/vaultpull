package sync

import (
	"fmt"
	"strings"
)

// DuplicateKeyError is returned when two secret paths produce the same env key.
type DuplicateKeyError struct {
	Key    string
	Paths  []string
}

func (e *DuplicateKeyError) Error() string {
	return fmt.Sprintf("duplicate env key %q produced by paths: %s",
		e.Key, strings.Join(e.Paths, ", "))
}

// Deduplicator tracks env keys and detects collisions across secret paths.
type Deduplicator struct {
	seen map[string]string // envKey -> original path
}

// NewDeduplicator creates a new Deduplicator.
func NewDeduplicator() *Deduplicator {
	return &Deduplicator{
		seen: make(map[string]string),
	}
}

// Check records the env key for the given path and returns a DuplicateKeyError
// if the key was already registered by a different path.
func (d *Deduplicator) Check(envKey, path string) error {
	if existing, ok := d.seen[envKey]; ok {
		if existing != path {
			return &DuplicateKeyError{
				Key:   envKey,
				Paths: []string{existing, path},
			}
		}
		return nil
	}
	d.seen[envKey] = path
	return nil
}

// Reset clears all recorded keys.
func (d *Deduplicator) Reset() {
	d.seen = make(map[string]string)
}

// Len returns the number of unique keys tracked.
func (d *Deduplicator) Len() int {
	return len(d.seen)
}
