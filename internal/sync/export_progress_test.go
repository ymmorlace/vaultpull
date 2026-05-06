package sync

import (
	"io"
)

// NewProgressReporterExported exposes NewProgressReporter for black-box tests.
func NewProgressReporterExported(out io.Writer, total int) *ProgressReporter {
	return NewProgressReporter(out, total)
}

// NewProgressWriterExported exposes NewProgressWriter for black-box tests.
func NewProgressWriterExported(inner EnvWriter, p *ProgressReporter) *ProgressWriter {
	return NewProgressWriter(inner, p)
}
