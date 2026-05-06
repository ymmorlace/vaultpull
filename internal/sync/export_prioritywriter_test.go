package sync

// Re-export priority constants for black-box tests.
const (
	PriorityHighExported   = PriorityHigh
	PriorityNormalExported = PriorityNormal
	PriorityLowExported    = PriorityLow
)

// NewPriorityWriterExported exposes NewPriorityWriter to external test packages.
func NewPriorityWriterExported(inner EnvWriter) *PriorityWriter {
	return NewPriorityWriter(inner)
}
