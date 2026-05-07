package sync_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	. "github.com/your-org/vaultpull/internal/sync"
)

func TestCheckpoint_RecordAndHas(t *testing.T) {
	cp := NewCheckpoint()
	if cp.Has("secret/app", "DB_PASS") {
		t.Fatal("expected miss on empty checkpoint")
	}
	cp.Record("secret/app", "DB_PASS")
	if !cp.Has("secret/app", "DB_PASS") {
		t.Fatal("expected hit after record")
	}
}

func TestCheckpoint_Reset_ClearsEntries(t *testing.T) {
	cp := NewCheckpoint()
	cp.Record("secret/app", "API_KEY")
	cp.Reset()
	if cp.Has("secret/app", "API_KEY") {
		t.Fatal("expected miss after reset")
	}
}

func TestCheckpoint_SaveAndLoad_RoundTrip(t *testing.T) {
	cp := NewCheckpoint()
	cp.Record("secret/app", "TOKEN")
	cp.Record("secret/db", "PASSWORD")

	tmp := filepath.Join(t.TempDir(), "checkpoint.json")
	if err := cp.Save(tmp); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := LoadCheckpoint(tmp)
	if err != nil {
		t.Fatalf("LoadCheckpoint: %v", err)
	}
	if !loaded.Has("secret/app", "TOKEN") {
		t.Error("expected secret/app TOKEN after load")
	}
	if !loaded.Has("secret/db", "PASSWORD") {
		t.Error("expected secret/db PASSWORD after load")
	}
}

func TestLoadCheckpoint_MissingFile(t *testing.T) {
	_, err := LoadCheckpoint(filepath.Join(t.TempDir(), "nope.json"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestCheckpoint_Save_CreatesFileWithRestrictedPerms(t *testing.T) {
	cp := NewCheckpoint()
	cp.Record("secret/x", "KEY")
	tmp := filepath.Join(t.TempDir(), "cp.json")
	if err := cp.Save(tmp); err != nil {
		t.Fatalf("Save: %v", err)
	}
	info, err := os.Stat(tmp)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("expected 0600, got %o", info.Mode().Perm())
	}
}

// --- CheckpointWriter tests ---

type recordingWriter struct {
	calls []struct{ path, key, value string }
	err   error
}

func (r *recordingWriter) Write(_ context.Context, path, key, value string) error {
	if r.err != nil {
		return r.err
	}
	r.calls = append(r.calls, struct{ path, key, value string }{path, key, value})
	return nil
}

func TestCheckpointWriter_SkipsAlreadyRecorded(t *testing.T) {
	cp := NewCheckpoint()
	cp.Record("secret/app", "SKIP_ME")
	rw := &recordingWriter{}
	w := NewCheckpointWriter(rw, cp)

	if err := w.Write(context.Background(), "secret/app", "SKIP_ME", "val"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rw.calls) != 0 {
		t.Errorf("expected 0 inner calls, got %d", len(rw.calls))
	}
}

func TestCheckpointWriter_RecordsOnSuccess(t *testing.T) {
	cp := NewCheckpoint()
	rw := &recordingWriter{}
	w := NewCheckpointWriter(rw, cp)

	if err := w.Write(context.Background(), "secret/app", "NEW_KEY", "v"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cp.Has("secret/app", "NEW_KEY") {
		t.Error("expected checkpoint to record entry after successful write")
	}
}

func TestCheckpointWriter_DoesNotRecordOnError(t *testing.T) {
	cp := NewCheckpoint()
	rw := &recordingWriter{err: errors.New("disk full")}
	w := NewCheckpointWriter(rw, cp)

	_ = w.Write(context.Background(), "secret/app", "FAIL_KEY", "v")
	if cp.Has("secret/app", "FAIL_KEY") {
		t.Error("checkpoint must not record entry when inner write fails")
	}
}
