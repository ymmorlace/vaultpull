package sync_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/your-org/vaultpull/internal/sync"
)

// stubWriter records calls and optionally returns an error.
type stubWriter struct {
	writes []struct{ path, key, value string }
	errOn  string // key that triggers an error
}

func (s *stubWriter) Write(path, key, value string) error {
	if key == s.errOn {
		return errors.New("stub write error")
	}
	s.writes = append(s.writes, struct{ path, key, value string }{path, key, value})
	return nil
}

func TestSnapshotWriter_RecordsEntries(t *testing.T) {
	inner := &stubWriter{}
	sw := sync.NewSnapshotWriter(inner)

	_ = sw.Write("secret/app", "DB_HOST", "localhost")
	_ = sw.Write("secret/app", "DB_PORT", "5432")

	snap := sw.Snapshot()
	if len(snap.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(snap.Entries))
	}
	if snap.Entries[0].Key != "DB_HOST" {
		t.Errorf("expected DB_HOST, got %s", snap.Entries[0].Key)
	}
	if snap.Entries[1].Value != "5432" {
		t.Errorf("expected 5432, got %s", snap.Entries[1].Value)
	}
}

func TestSnapshotWriter_DoesNotRecordOnError(t *testing.T) {
	inner := &stubWriter{errOn: "BAD_KEY"}
	sw := sync.NewSnapshotWriter(inner)

	_ = sw.Write("secret/app", "GOOD_KEY", "val")
	err := sw.Write("secret/app", "BAD_KEY", "val")

	if err == nil {
		t.Fatal("expected error from inner writer")
	}
	snap := sw.Snapshot()
	if len(snap.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(snap.Entries))
	}
}

func TestSnapshotWriter_TimestampIsSet(t *testing.T) {
	sw := sync.NewSnapshotWriter(&stubWriter{})
	snap := sw.Snapshot()
	if snap.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestSaveSnapshot_WritesValidJSON(t *testing.T) {
	sw := sync.NewSnapshotWriter(&stubWriter{})
	_ = sw.Write("secret/app", "API_KEY", "abc123")

	tmp := filepath.Join(t.TempDir(), "snapshot.json")
	if err := sync.SaveSnapshot(sw.Snapshot(), tmp); err != nil {
		t.Fatalf("SaveSnapshot error: %v", err)
	}

	data, err := os.ReadFile(tmp)
	if err != nil {
		t.Fatalf("read snapshot file: %v", err)
	}
	var snap sync.Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		t.Fatalf("unmarshal snapshot: %v", err)
	}
	if len(snap.Entries) != 1 || snap.Entries[0].Key != "API_KEY" {
		t.Errorf("unexpected snapshot contents: %+v", snap)
	}
}

func TestSaveSnapshot_InvalidPath(t *testing.T) {
	sw := sync.NewSnapshotWriter(&stubWriter{})
	err := sync.SaveSnapshot(sw.Snapshot(), "/nonexistent/dir/snap.json")
	if err == nil {
		t.Fatal("expected error for invalid path")
	}
}
