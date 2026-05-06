package sync_test

import (
	"context"
	"errors"
	"testing"

	"github.com/user/vaultpull/internal/sync"
)

func TestNoEmptyKey_RejectsBlank(t *testing.T) {
	err := sync.NoEmptyKey("", "value")
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestNoEmptyKey_AcceptsNonBlank(t *testing.T) {
	if err := sync.NoEmptyKey("MY_KEY", "value"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNoEmptyValue_RejectsBlank(t *testing.T) {
	err := sync.NoEmptyValue("KEY", "   ")
	if err == nil {
		t.Fatal("expected error for blank value")
	}
}

func TestNoEmptyValue_AcceptsNonBlank(t *testing.T) {
	if err := sync.NoEmptyValue("KEY", "secret"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMaxValueLength_Exceeds(t *testing.T) {
	rule := sync.MaxValueLength(5)
	if err := rule("KEY", "toolong"); err == nil {
		t.Fatal("expected error for value exceeding max length")
	}
}

func TestMaxValueLength_WithinLimit(t *testing.T) {
	rule := sync.MaxValueLength(10)
	if err := rule("KEY", "ok"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidator_FirstErrorReturned(t *testing.T) {
	v := sync.NewValidator(sync.NoEmptyKey, sync.NoEmptyValue)
	err := v.Validate("", "")
	if err == nil {
		t.Fatal("expected error")
	}
	var ve *sync.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %T", err)
	}
}

func TestValidatingWriter_BlocksInvalidEntry(t *testing.T) {
	captured := &captureWriter{}
	v := sync.NewValidator(sync.NoEmptyKey)
	w := sync.NewValidatingWriter(captured, v)

	err := w.Write(context.Background(), "", "value")
	if err == nil {
		t.Fatal("expected validation error")
	}
	if len(captured.entries) != 0 {
		t.Fatalf("inner writer should not have been called, got %d entries", len(captured.entries))
	}
}

func TestValidatingWriter_PassesValidEntry(t *testing.T) {
	captured := &captureWriter{}
	v := sync.NewValidator(sync.NoEmptyKey, sync.NoEmptyValue)
	w := sync.NewValidatingWriter(captured, v)

	if err := w.Write(context.Background(), "API_KEY", "abc123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(captured.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(captured.entries))
	}
}

func TestValidatingWriter_PanicOnNilInner(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner writer")
		}
	}()
	sync.NewValidatingWriter(nil, sync.NewValidator())
}

func TestValidatingWriter_PanicOnNilValidator(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil validator")
		}
	}()
	sync.NewValidatingWriter(&captureWriter{}, nil)
}
