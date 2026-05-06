package sync

import (
	"io"
)

// NewDryRunWriterExported exposes NewDryRunWriter for black-box tests
// that live in package sync_test but need access to the constructor
// without importing the full package path.
func NewDryRunWriterExported(out io.Writer) *DryRunWriter {
	return NewDryRunWriter(out)
}
