package sync

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// captureWriter records every Write call.
type captureWriter struct {
	entries []string
	writeErr error
}

func (c *captureWriter) Write(_ context.Context, key, value string) error {
	if c.writeErr != nil {
		return c.writeErr
	}
	c.entries = append(c.entries, key+"="+value)
	return nil
}

func TestSplitter_PanicOnNoWriters(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	NewSplitter(func(string) int { return 0 })
}

func TestSplitter_PanicOnNilRouteFn(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	NewSplitter(nil, &captureWriter{})
}

func TestSplitter_RoutesToCorrectWriter(t *testing.T) {
	a, b := &captureWriter{}, &captureWriter{}
	s := NewSplitter(func(key string) int {
		if strings.HasPrefix(key, "APP_") {
			return 0
		}
		return 1
	}, a, b)

	ctx := context.Background()
	_ = s.Write(ctx, "APP_SECRET", "x")
	_ = s.Write(ctx, "DB_PASS", "y")

	if len(a.entries) != 1 || a.entries[0] != "APP_SECRET=x" {
		t.Errorf("writer a: got %v", a.entries)
	}
	if len(b.entries) != 1 || b.entries[0] != "DB_PASS=y" {
		t.Errorf("writer b: got %v", b.entries)
	}
}

func TestSplitter_BroadcastOnMinusOne(t *testing.T) {
	a, b := &captureWriter{}, &captureWriter{}
	s := NewSplitter(func(string) int { return -1 }, a, b)

	_ = s.Write(context.Background(), "KEY", "val")

	if len(a.entries) != 1 || len(b.entries) != 1 {
		t.Errorf("expected both writers to receive entry; a=%v b=%v", a.entries, b.entries)
	}
}

func TestSplitter_OutOfRangeIndexReturnsError(t *testing.T) {
	s := NewSplitter(func(string) int { return 99 }, &captureWriter{})
	err := s.Write(context.Background(), "K", "V")
	if err == nil {
		t.Fatal("expected error for out-of-range index")
	}
}

func TestSplitter_PropagatesInnerError(t *testing.T) {
	sentinel := errors.New("write failed")
	w := &captureWriter{writeErr: sentinel}
	s := NewSplitter(func(string) int { return 0 }, w)

	err := s.Write(context.Background(), "K", "V")
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestSplitter_Len(t *testing.T) {
	s := NewSplitter(func(string) int { return 0 }, &captureWriter{}, &captureWriter{}, &captureWriter{})
	if s.Len() != 3 {
		t.Fatalf("expected 3, got %d", s.Len())
	}
}
