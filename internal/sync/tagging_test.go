package sync_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultpull/internal/sync"
)

// captureWriter records Write calls for assertion.
type captureWriter struct {
	entries []struct{ key, value string }
	errOn  string // return error when key equals this
}

func (c *captureWriter) Write(_ context.Context, key, value string) error {
	if c.errOn != "" && key == c.errOn {
		return errors.New("capture: forced error")
	}
	c.entries = append(c.entries, struct{ key, value string }{key, value})
	return nil
}

func TestTagger_PanicOnNilInner(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner writer")
		}
	}()
	sync.NewTagger(nil, "", sync.Tag{Key: "env", Value: "test"})
}

func TestTagger_NoTags_PassesThrough(t *testing.T) {
	cw := &captureWriter{}
	tg := sync.NewTagger(cw, "")
	if err := tg.Write(context.Background(), "FOO", "bar"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cw.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(cw.entries))
	}
	if cw.entries[0].key != "FOO" || cw.entries[0].value != "bar" {
		t.Errorf("unexpected entry: %+v", cw.entries[0])
	}
}

func TestTagger_WritesPrefixComments(t *testing.T) {
	cw := &captureWriter{}
	tg := sync.NewTagger(cw, "# %s=%s",
		sync.Tag{Key: "source", Value: "vault"},
		sync.Tag{Key: "env", Value: "prod"},
	)
	if err := tg.Write(context.Background(), "DB_PASS", "secret"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 2 tag comments + 1 real entry
	if len(cw.entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(cw.entries))
	}
	if cw.entries[2].key != "DB_PASS" {
		t.Errorf("last entry should be real key, got %q", cw.entries[2].key)
	}
}

func TestTagger_WithTag_IsImmutable(t *testing.T) {
	cw := &captureWriter{}
	base := sync.NewTagger(cw, "", sync.Tag{Key: "a", Value: "1"})
	extended := base.WithTag("b", "2")

	if len(base.Tags()) != 1 {
		t.Errorf("base should still have 1 tag, got %d", len(base.Tags()))
	}
	if len(extended.Tags()) != 2 {
		t.Errorf("extended should have 2 tags, got %d", len(extended.Tags()))
	}
}

func TestTagger_InnerError_PropagatesOnTagWrite(t *testing.T) {
	cw := &captureWriter{errOn: "#source"}
	tg := sync.NewTagger(cw, "# %s=%s", sync.Tag{Key: "source", Value: "vault"})
	err := tg.Write(context.Background(), "KEY", "val")
	if err == nil {
		t.Fatal("expected error when inner writer fails on tag")
	}
}
