package sync

import (
	"context"
	"fmt"
	"strings"
)

// Tag represents a key-value metadata pair attached to a secret entry.
type Tag struct {
	Key   string
	Value string
}

// Tagger enriches secret entries with metadata tags before passing them
// to an inner EnvWriter. Tags are appended as comments above each entry.
type Tagger struct {
	inner  EnvWriter
	tags   []Tag
	format string // e.g. "# %s=%s"
}

// NewTagger wraps inner with a tagger that prepends comment lines for each
// tag. format must contain exactly two %s verbs (key, value).
func NewTagger(inner EnvWriter, format string, tags ...Tag) *Tagger {
	if inner == nil {
		panic("tagger: inner writer must not be nil")
	}
	if format == "" {
		format = "# %s=%s"
	}
	return &Tagger{inner: inner, tags: tags, format: format}
}

// Write emits one comment line per tag then delegates to the inner writer.
func (t *Tagger) Write(ctx context.Context, key, value string) error {
	for _, tag := range t.tags {
		comment := fmt.Sprintf(t.format, tag.Key, tag.Value)
		if err := t.inner.Write(ctx, "#"+sanitizeTagKey(tag.Key), comment); err != nil {
			return fmt.Errorf("tagger: writing tag %q: %w", tag.Key, err)
		}
	}
	return t.inner.Write(ctx, key, value)
}

// WithTag returns a new Tagger with the additional tag appended.
func (t *Tagger) WithTag(key, value string) *Tagger {
	newTags := make([]Tag, len(t.tags)+1)
	copy(newTags, t.tags)
	newTags[len(t.tags)] = Tag{Key: key, Value: value}
	return &Tagger{inner: t.inner, tags: newTags, format: t.format}
}

// Tags returns a copy of the current tag set.
func (t *Tagger) Tags() []Tag {
	out := make([]Tag, len(t.tags))
	copy(out, t.tags)
	return out
}

func sanitizeTagKey(k string) string {
	return strings.Map(func(r rune) rune {
		if r == ' ' || r == '\t' {
			return '_'
		}
		return r
	}, k)
}
