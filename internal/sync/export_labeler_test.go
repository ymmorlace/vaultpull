package sync

// NewLabelerExported re-exports NewLabeler for white-box tests in the sync
// package that need access to unexported helpers.
func NewLabelerExported(inner EnvWriter, prefix string, labels ...Label) *Labeler {
	return NewLabeler(inner, prefix, labels...)
}

// ApplyLabelsExported exposes the unexported applyLabels method for unit tests.
func ApplyLabelsExported(l *Labeler, key string) string {
	return l.applyLabels(key)
}
