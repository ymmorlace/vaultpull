package sync

// Re-export Tag so external test package can reference it without import alias issues.
type TagExported = Tag

// NewTaggerExported exposes NewTagger for black-box tests.
func NewTaggerExported(inner EnvWriter, format string, tags ...Tag) *Tagger {
	return NewTagger(inner, format, tags...)
}
