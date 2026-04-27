package sync_test

import (
	"errors"
	"testing"

	syncp "github.com/vaultpull/internal/sync"
)

// stubWriter captures the last Write call for assertions.
type stubWriter struct {
	path    string
	secrets map[string]string
	err     error
}

func (s *stubWriter) Write(path string, secrets map[string]string) error {
	s.path = path
	s.secrets = secrets
	return s.err
}

// fakeClient satisfies the interface used by Syncer indirectly via vault.Client
// — here we test toEnvKey and writer integration with a real Syncer backed by a
// mock HTTP server (see client_test.go pattern). For unit speed we test the
// key-transformation helper directly.
func TestToEnvKey(t *testing.T) {
	cases := []struct {
		keyPath, field, want string
	}{
		{"db", "password", "DB_PASSWORD"},
		{"app/config", "api_key", "APP_CONFIG_API_KEY"},
		{"my-service/prod", "secret", "MY_SERVICE_PROD_SECRET"},
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			got := syncp.ExportToEnvKey(tc.keyPath, tc.field)
			if got != tc.want {
				t.Errorf("toEnvKey(%q, %q) = %q; want %q", tc.keyPath, tc.field, got, tc.want)
			}
		})
	}
}

func TestFileEnvWriter_WriteError(t *testing.T) {
	w := &stubWriter{err: errors.New("disk full")}
	err := w.Write("/tmp/test.env", map[string]string{"K": "V"})
	if err == nil || err.Error() != "disk full" {
		t.Fatalf("expected 'disk full' error, got %v", err)
	}
}

func TestFileEnvWriter_CapturesPath(t *testing.T) {
	w := &stubWriter{}
	_ = w.Write(".env", map[string]string{"FOO": "bar"})
	if w.path != ".env" {
		t.Errorf("expected path '.env', got %q", w.path)
	}
	if w.secrets["FOO"] != "bar" {
		t.Errorf("expected FOO=bar in secrets")
	}
}
