package sync_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	syncp "github.com/vaultpull/internal/sync"
)

func TestFileEnvWriter_Write(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, ".env")

	w := syncp.FileEnvWriter{}
	secrets := map[string]string{
		"DB_PASSWORD": "s3cr3t",
		"APP_PORT":    "8080",
		"GREETING":    "hello world",
	}

	if err := w.Write(outPath, secrets); err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}

	content := string(data)
	for _, want := range []string{"APP_PORT=8080", "DB_PASSWORD=s3cr3t"} {
		if !strings.Contains(content, want) {
			t.Errorf("expected %q in output:\n%s", want, content)
		}
	}

	// Values with spaces must be quoted.
	if !strings.Contains(content, `GREETING=`) {
		t.Errorf("expected GREETING key in output")
	}
	if strings.Contains(content, "GREETING=hello world\n") {
		t.Errorf("value with space should be quoted")
	}
}

func TestFileEnvWriter_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, ".env")

	w := syncp.FileEnvWriter{}
	if err := w.Write(outPath, map[string]string{"K": "V"}); err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	info, err := os.Stat(outPath)
	if err != nil {
		t.Fatalf("Stat() error: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("expected file permissions 0600, got %o", perm)
	}
}
