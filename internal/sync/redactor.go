package sync

import (
	"regexp"
	"strings"
)

// RedactPattern describes a named pattern whose matching values should be redacted.
type RedactPattern struct {
	Name    string
	Pattern *regexp.Regexp
}

// Redactor masks sensitive secret values before they are written or logged.
type Redactor struct {
	patterns  []RedactPattern
	mask      string
}

// NewRedactor returns a Redactor that replaces values matching any of the
// supplied patterns with mask. If mask is empty, "***" is used.
func NewRedactor(mask string, patterns ...RedactPattern) *Redactor {
	if mask == "" {
		mask = "***"
	}
	return &Redactor{
		patterns: patterns,
		mask:     mask,
	}
}

// Redact returns the masked string if value matches any registered pattern,
// otherwise it returns value unchanged.
func (r *Redactor) Redact(key, value string) string {
	for _, p := range r.patterns {
		if p.Pattern.MatchString(value) {
			return r.mask
		}
	}
	return value
}

// IsSensitiveKey returns true when the key name suggests it holds a secret
// (e.g. contains "password", "token", "secret", "key", "apikey").
func IsSensitiveKey(key string) bool {
	lower := strings.ToLower(key)
	for _, kw := range []string{"password", "passwd", "token", "secret", "apikey", "api_key", "private"} {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// DefaultSensitivePatterns returns a set of RedactPatterns covering common
// high-entropy and structured secret formats.
func DefaultSensitivePatterns() []RedactPattern {
	return []RedactPattern{
		{
			Name:    "jwt",
			Pattern: regexp.MustCompile(`^ey[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+$`),
		},
		{
			Name:    "aws-secret",
			Pattern: regexp.MustCompile(`^[A-Za-z0-9/+=]{40}$`),
		},
	}
}
