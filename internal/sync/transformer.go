package sync

import (
	"fmt"
	"strings"
)

// ValueTransformer applies transformations to secret values before writing.
type ValueTransformer interface {
	Transform(key, value string) (string, error)
}

// ChainTransformer applies multiple transformers in sequence.
type ChainTransformer struct {
	transformers []ValueTransformer
}

// NewChainTransformer creates a transformer that applies each transformer in order.
func NewChainTransformer(transformers ...ValueTransformer) *ChainTransformer {
	return &ChainTransformer{transformers: transformers}
}

// Transform applies all transformers in sequence.
func (c *ChainTransformer) Transform(key, value string) (string, error) {
	var err error
	for _, t := range c.transformers {
		value, err = t.Transform(key, value)
		if err != nil {
			return "", fmt.Errorf("transformer error for key %q: %w", key, err)
		}
	}
	return value, nil
}

// TrimSpaceTransformer removes leading and trailing whitespace from values.
type TrimSpaceTransformer struct{}

// NewTrimSpaceTransformer creates a new TrimSpaceTransformer.
func NewTrimSpaceTransformer() *TrimSpaceTransformer {
	return &TrimSpaceTransformer{}
}

// Transform trims whitespace from the value.
func (t *TrimSpaceTransformer) Transform(_, value string) (string, error) {
	return strings.TrimSpace(value), nil
}

// QuoteTransformer wraps values containing special characters in double quotes.
type QuoteTransformer struct{}

// NewQuoteTransformer creates a new QuoteTransformer.
func NewQuoteTransformer() *QuoteTransformer {
	return &QuoteTransformer{}
}

// Transform quotes the value if it contains spaces, newlines, or hash characters.
func (q *QuoteTransformer) Transform(_, value string) (string, error) {
	if strings.ContainsAny(value, " \t\n\r#") {
		escaped := strings.ReplaceAll(value, `"`, `\"`)
		return fmt.Sprintf(`"%s"`, escaped), nil
	}
	return value, nil
}

// PrefixTransformer prepends a fixed string to every secret value.
type PrefixTransformer struct {
	prefix string
}

// NewPrefixTransformer creates a new PrefixTransformer that prepends the given prefix.
func NewPrefixTransformer(prefix string) *PrefixTransformer {
	return &PrefixTransformer{prefix: prefix}
}

// Transform prepends the configured prefix to the value.
func (p *PrefixTransformer) Transform(_, value string) (string, error) {
	return p.prefix + value, nil
}

// SuffixTransformer appends a fixed string to every secret value.
type SuffixTransformer struct {
	suffix string
}

// NewSuffixTransformer creates a new SuffixTransformer that appends the given suffix.
func NewSuffixTransformer(suffix string) *SuffixTransformer {
	return &SuffixTransformer{suffix: suffix}
}

// Transform appends the configured suffix to the value.
func (s *SuffixTransformer) Transform(_, value string) (string, error) {
	return value + s.suffix, nil
}
