package sync_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultpull/internal/sync"
)

// captureWriter records the last key/value pair written to it.
type captureWriter struct {
	key   string
	value string
	err   error
}

func (c *captureWriter) Write(_ context.Context, key, value string) error {
	if c.err != nil {
		return c.err
	}
	c.key = key
	c.value = value
	return nil
}

func TestEnvelopeWriter_PanicOnNilInner(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner writer")
		}
	}()
	sync.NewEnvelopeWriter(nil, "PRE", "SUF", "__")
}

func TestEnvelopeWriter_PanicOnEmptySep(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty separator")
		}
	}()
	sync.NewEnvelopeWriter(&captureWriter{}, "PRE", "SUF", "")
}

func TestEnvelopeWriter_PrefixAndSuffix(t *testing.T) {
	cap := &captureWriter{}
	w := sync.NewEnvelopeWriter(cap, "APP", "V1", "__")

	if err := w.Write(context.Background(), "DB_HOST", "localhost"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cap.key != "APP__DB_HOST__V1" {
		t.Errorf("got key %q, want %q", cap.key, "APP__DB_HOST__V1")
	}
	if cap.value != "localhost" {
		t.Errorf("got value %q, want %q", cap.value, "localhost")
	}
}

func TestEnvelopeWriter_PrefixOnly(t *testing.T) {
	cap := &captureWriter{}
	w := sync.NewEnvelopeWriter(cap, "NS", "", "_")

	_ = w.Write(context.Background(), "TOKEN", "secret")

	if cap.key != "NS_TOKEN" {
		t.Errorf("got key %q, want %q", cap.key, "NS_TOKEN")
	}
}

func TestEnvelopeWriter_SuffixOnly(t *testing.T) {
	cap := &captureWriter{}
	w := sync.NewEnvelopeWriter(cap, "", "PROD", "-")

	_ = w.Write(context.Background(), "API_KEY", "xyz")

	if cap.key != "API_KEY-PROD" {
		t.Errorf("got key %q, want %q", cap.key, "API_KEY-PROD")
	}
}

func TestEnvelopeWriter_NoPrefixNoSuffix(t *testing.T) {
	cap := &captureWriter{}
	w := sync.NewEnvelopeWriter(cap, "", "", "__")

	_ = w.Write(context.Background(), "PLAIN", "val")

	if cap.key != "PLAIN" {
		t.Errorf("got key %q, want %q", cap.key, "PLAIN")
	}
}

func TestEnvelopeWriter_PropagatesInnerError(t *testing.T) {
	expected := errors.New("write failed")
	cap := &captureWriter{err: expected}
	w := sync.NewEnvelopeWriter(cap, "PRE", "", "_")

	err := w.Write(context.Background(), "KEY", "val")
	if !errors.Is(err, expected) {
		t.Errorf("got %v, want %v", err, expected)
	}
}

func TestEnvelopeWriter_Describe(t *testing.T) {
	w := sync.NewEnvelopeWriter(&captureWriter{}, "APP", "V2", "__")
	desc := w.Describe()
	if desc == "" {
		t.Error("expected non-empty description")
	}
}
