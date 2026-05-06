package sync_test

import (
	"regexp"
	"testing"

	"github.com/yourusername/vaultpull/internal/sync"
)

func TestRedactor_NoPatterns_ReturnsOriginal(t *testing.T) {
	r := sync.NewRedactor("") // no patterns
	got := r.Redact("KEY", "plainvalue")
	if got != "plainvalue" {
		t.Errorf("expected plainvalue, got %q", got)
	}
}

func TestRedactor_MatchingPattern_ReturnsMask(t *testing.T) {
	p := sync.RedactPattern{
		Name:    "digits",
		Pattern: regexp.MustCompile(`^\d+$`),
	}
	r := sync.NewRedactor("REDACTED", p)
	got := r.Redact("PORT", "9200")
	if got != "REDACTED" {
		t.Errorf("expected REDACTED, got %q", got)
	}
}

func TestRedactor_NonMatchingPattern_ReturnsOriginal(t *testing.T) {
	p := sync.RedactPattern{
		Name:    "digits",
		Pattern: regexp.MustCompile(`^\d+$`),
	}
	r := sync.NewRedactor("***", p)
	got := r.Redact("HOST", "localhost")
	if got != "localhost" {
		t.Errorf("expected localhost, got %q", got)
	}
}

func TestRedactor_DefaultMask(t *testing.T) {
	p := sync.RedactPattern{
		Name:    "all",
		Pattern: regexp.MustCompile(`.+`),
	}
	r := sync.NewRedactor("", p) // empty mask → "***"
	got := r.Redact("X", "anything")
	if got != "***" {
		t.Errorf("expected ***, got %q", got)
	}
}

func TestIsSensitiveKey(t *testing.T) {
	cases := []struct {
		key       string
		expected  bool
	}{
		{"DB_PASSWORD", true},
		{"AUTH_TOKEN", true},
		{"API_SECRET", true},
		{"PRIVATE_KEY", true},
		{"DB_HOST", false},
		{"PORT", false},
		{"APP_NAME", false},
	}
	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			got := sync.IsSensitiveKey(tc.key)
			if got != tc.expected {
				t.Errorf("IsSensitiveKey(%q) = %v, want %v", tc.key, got, tc.expected)
			}
		})
	}
}

func TestDefaultSensitivePatterns_JWT(t *testing.T) {
	patterns := sync.DefaultSensitivePatterns()
	r := sync.NewRedactor("***", patterns...)
	jwt := "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJ1c2VyIn0.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
	got := r.Redact("TOKEN", jwt)
	if got != "***" {
		t.Errorf("expected JWT to be redacted, got %q", got)
	}
}

func TestDefaultSensitivePatterns_PlainValue_NotRedacted(t *testing.T) {
	patterns := sync.DefaultSensitivePatterns()
	r := sync.NewRedactor("***", patterns...)
	got := r.Redact("APP_ENV", "production")
	if got != "production" {
		t.Errorf("expected production to pass through, got %q", got)
	}
}
