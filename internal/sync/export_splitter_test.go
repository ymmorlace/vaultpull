package sync

// NewSplitterExported exposes NewSplitter for black-box tests in the sync_test package.
func NewSplitterExported(routeFn RouteFn, writers ...splitterWriter) *Splitter {
	return NewSplitter(routeFn, writers...)
}

// CaptureWriterExported is a test helper that records writes.
type CaptureWriterExported = captureWriter
