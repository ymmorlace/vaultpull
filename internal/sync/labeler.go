package sync

import (
	"fmt"
	"strings"
)

// Label represents a key-value metadata tag attached to a secret entry.
type Label struct {
	Key   string
	Value string
}

// Labeler attaches static or dynamic labels to secret keys before writing.
type Labeler struct {
	inner  EnvWriter
	labels []Label
	prefix string
}

// NewLabeler wraps an EnvWriter and injects label-derived comments or prefixes
// into the key name. Labels are applied in order; conflicting keys are suffixed
// with the label value separated by an underscore.
//
// prefix is prepended to every key (e.g. "APP"). Pass an empty string to skip.
func NewLabeler(inner EnvWriter, prefix string, labels ...Label) *Labeler {
	if inner == nil {
		panic("labeler: inner writer must not be nil")
	}
	return &Labeler{
		inner:  inner,
		labels: labels,
		prefix: strings.ToUpper(strings.TrimSpace(prefix)),
	}
}

// Write applies the configured prefix and labels to key, then delegates to the
// inner writer.
func (l *Labeler) Write(ctx interface{ Done() <-chan struct{} }, key, value string) error {
	labeled := l.applyLabels(key)
	return l.inner.Write(ctx, labeled, value)
}

// applyLabels builds the final key from prefix and label values.
func (l *Labeler) applyLabels(key string) string {
	parts := make([]string, 0, 1+len(l.labels)+1)
	if l.prefix != "" {
		parts = append(parts, l.prefix)
	}
	for _, lbl := range l.labels {
		v := strings.ToUpper(strings.TrimSpace(lbl.Value))
		v = strings.NewReplacer("-", "_", ".", "_", " ", "_").Replace(v)
		if v != "" {
			parts = append(parts, v)
		}
	}
	parts = append(parts, key)
	return strings.Join(parts, "_")
}

// Labels returns a copy of the configured labels.
func (l *Labeler) Labels() []Label {
	out := make([]Label, len(l.labels))
	copy(out, l.labels)
	return out
}

// String returns a human-readable description of the labeler configuration.
func (l *Labeler) String() string {
	return fmt.Sprintf("Labeler{prefix=%q labels=%d}", l.prefix, len(l.labels))
}
