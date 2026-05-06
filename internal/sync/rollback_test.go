package sync_test

import (
	"context"
	"errors"
	"testing"

	"github.com/user/vaultpull/internal/sync"
)

func TestRollbackWriter_PanicOnNilInner(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner writer")
		}
	}()
	sync.NewRollbackWriter(nil, nil)
}

func TestRollbackWriter_RecordsPreviousValue(t *testing.T) {
	buf := &captureWriter{}
	prev := map[string]string{"DB_HOST": "localhost"}
	rw := sync.NewRollbackWriter(buf, prev)

	if err := rw.Write(context.Background(), "DB_HOST", "prod.db"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries := rw.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Key != "DB_HOST" {
		t.Errorf("expected key DB_HOST, got %s", entries[0].Key)
	}
	if entries[0].OldValue != "localhost" {
		t.Errorf("expected old value localhost, got %s", entries[0].OldValue)
	}
	if !entries[0].HadValue {
		t.Error("expected HadValue=true for existing key")
	}
}

func TestRollbackWriter_RecordsAbsentKey(t *testing.T) {
	buf := &captureWriter{}
	rw := sync.NewRollbackWriter(buf, nil)

	_ = rw.Write(context.Background(), "NEW_KEY", "value")

	entries := rw.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].HadValue {
		t.Error("expected HadValue=false for absent key")
	}
	if entries[0].OldValue != "" {
		t.Errorf("expected empty old value, got %q", entries[0].OldValue)
	}
}

func TestRollbackWriter_PropagatesInnerError(t *testing.T) {
	sentinel := errors.New("write failed")
	ew := &errorWriter{err: sentinel}
	rw := sync.NewRollbackWriter(ew, nil)

	err := rw.Write(context.Background(), "KEY", "val")
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestRollbackWriter_Reset_ClearsLog(t *testing.T) {
	buf := &captureWriter{}
	rw := sync.NewRollbackWriter(buf, nil)

	_ = rw.Write(context.Background(), "A", "1")
	_ = rw.Write(context.Background(), "B", "2")
	rw.Reset()

	if len(rw.Entries()) != 0 {
		t.Errorf("expected empty log after reset, got %d entries", len(rw.Entries()))
	}
}

func TestRollbackWriter_MultipleWrites_AllRecorded(t *testing.T) {
	buf := &captureWriter{}
	rw := sync.NewRollbackWriter(buf, map[string]string{"X": "old"})

	keys := []string{"X", "Y", "Z"}
	for _, k := range keys {
		_ = rw.Write(context.Background(), k, "new")
	}

	if got := len(rw.Entries()); got != 3 {
		t.Errorf("expected 3 entries, got %d", got)
	}
}
