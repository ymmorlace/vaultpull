package sync

// NewEnvelopeWriterExported exposes NewEnvelopeWriter to external test packages
// without requiring them to import internal types directly.
func NewEnvelopeWriterExported(inner EnvWriter, prefix, suffix, sep string) *EnvelopeWriter {
	return NewEnvelopeWriter(inner, prefix, suffix, sep)
}
