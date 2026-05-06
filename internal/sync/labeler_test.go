package sync_test

import (
	"context"
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
	c.key = key
	c.value = value
	return c.err
}

func TestLabeler_PanicOnNilInner(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner writer")
		}
	}()
	sync.NewLabeler(nil, "APP")
}

func TestLabeler_NoLabelsNoPrefix(t *testing.T) {
	cw := &captureWriter{}
	l := sync.NewLabeler(cw, "")
	if err := l.Write(context.Background(), "DB_HOST", "localhost"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cw.key != "DB_HOST" {
		t.Errorf("key = %q; want %q", cw.key, "DB_HOST")
	}
}

func TestLabeler_PrefixPrependedToKey(t *testing.T) {
	cw := &captureWriter{}
	l := sync.NewLabeler(cw, "myapp")
	_ = l.Write(context.Background(), "PORT", "8080")
	if cw.key != "MYAPP_PORT" {
		t.Errorf("key = %q; want %q", cw.key, "MYAPP_PORT")
	}
}

func TestLabeler_LabelInsertedBetweenPrefixAndKey(t *testing.T) {
	cw := &captureWriter{}
	l := sync.NewLabeler(cw, "app", sync.Label{Key: "env", Value: "production"})
	_ = l.Write(context.Background(), "SECRET", "val")
	if cw.key != "APP_PRODUCTION_SECRET" {
		t.Errorf("key = %q; want %q", cw.key, "APP_PRODUCTION_SECRET")
	}
}

func TestLabeler_LabelValueSanitised(t *testing.T) {
	cw := &captureWriter{}
	l := sync.NewLabeler(cw, "", sync.Label{Key: "region", Value: "us-east.1"})
	_ = l.Write(context.Background(), "KEY", "v")
	if cw.key != "US_EAST_1_KEY" {
		t.Errorf("key = %q; want %q", cw.key, "US_EAST_1_KEY")
	}
}

func TestLabeler_MultipleLabels(t *testing.T) {
	cw := &captureWriter{}
	l := sync.NewLabeler(cw, "svc",
		sync.Label{Key: "env", Value: "staging"},
		sync.Label{Key: "tier", Value: "backend"},
	)
	_ = l.Write(context.Background(), "TOKEN", "abc")
	if cw.key != "SVC_STAGING_BACKEND_TOKEN" {
		t.Errorf("key = %q; want %q", cw.key, "SVC_STAGING_BACKEND_TOKEN")
	}
}

func TestLabeler_PropagatesInnerError(t *testing.T) {
	cw := &captureWriter{err: fmt.Errorf("disk full")}
	l := sync.NewLabeler(cw, "")
	if err := l.Write(context.Background(), "K", "V"); err == nil {
		t.Fatal("expected error from inner writer")
	}
}

func TestLabeler_Labels_ReturnsCopy(t *testing.T) {
	cw := &captureWriter{}
	labels := []sync.Label{{Key: "a", Value: "b"}}
	l := sync.NewLabeler(cw, "", labels...)
	copy := l.Labels()
	copy[0].Value = "mutated"
	if l.Labels()[0].Value != "b" {
		t.Error("Labels() should return a defensive copy")
	}
}
