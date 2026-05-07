package sync

import "github.com/yourusername/vaultpull/internal/sync"

// NewBufferedWriterExported re-exports the constructor for black-box tests.
func NewBufferedWriterExported(inner sync.EnvWriter, capacity int) *sync.BufferedWriter {
	return sync.NewBufferedWriter(inner, capacity)
}
