package sync

import (
	"errors"
	"testing"
)

func TestDeduplicator_NoDuplicates(t *testing.T) {
	d := NewDeduplicator()

	if err := d.Check("DB_HOST", "database/host"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := d.Check("DB_PORT", "database/port"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Len() != 2 {
		t.Fatalf("expected 2 keys, got %d", d.Len())
	}
}

func TestDeduplicator_SamePathSameKey(t *testing.T) {
	d := NewDeduplicator()

	if err := d.Check("DB_HOST", "database/host"); err != nil {
		t.Fatalf("unexpected error on first check: %v", err)
	}
	// Same key, same path — should be idempotent.
	if err := d.Check("DB_HOST", "database/host"); err != nil {
		t.Fatalf("unexpected error on repeat check: %v", err)
	}
}

func TestDeduplicator_DuplicateKey(t *testing.T) {
	d := NewDeduplicator()

	if err := d.Check("APP_SECRET", "app/secret"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err := d.Check("APP_SECRET", "application/secret")
	if err == nil {
		t.Fatal("expected duplicate key error, got nil")
	}

	var dupErr *DuplicateKeyError
	if !errors.As(err, &dupErr) {
		t.Fatalf("expected *DuplicateKeyError, got %T", err)
	}
	if dupErr.Key != "APP_SECRET" {
		t.Errorf("expected key APP_SECRET, got %q", dupErr.Key)
	}
	if len(dupErr.Paths) != 2 {
		t.Errorf("expected 2 paths, got %d", len(dupErr.Paths))
	}
}

func TestDeduplicator_ErrorMessage(t *testing.T) {
	err := &DuplicateKeyError{
		Key:   "FOO_BAR",
		Paths: []string{"foo/bar", "foo-bar"},
	}
	msg := err.Error()
	if msg == "" {
		t.Fatal("expected non-empty error message")
	}
	if !contains(msg, "FOO_BAR") {
		t.Errorf("error message missing key name: %s", msg)
	}
}

func TestDeduplicator_Reset(t *testing.T) {
	d := NewDeduplicator()
	_ = d.Check("KEY_ONE", "path/one")
	_ = d.Check("KEY_TWO", "path/two")

	d.Reset()
	if d.Len() != 0 {
		t.Fatalf("expected 0 keys after reset, got %d", d.Len())
	}
	// Should be able to re-register without collision.
	if err := d.Check("KEY_ONE", "path/other"); err != nil {
		t.Fatalf("unexpected error after reset: %v", err)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub ||
		len(s) > 0 && containsHelper(s, sub))
}

func containsHelper(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
