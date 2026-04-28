package sync_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/user/vaultpull/internal/sync"
	"github.com/user/vaultpull/internal/vault"
)

// stubWriter captures written secrets for assertions.
type stubWriter struct {
	data map[string]string
	err  error
}

func (s *stubWriter) Write(secrets map[string]string) error {
	if s.err != nil {
		return s.err
	}
	s.data = secrets
	return nil
}

func newRunnerServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "metadata") && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{"keys":["app/db_password","app/api_key"]}}`))
		case strings.Contains(r.URL.Path, "data"):
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":{"data":{"value":"secret123"}}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestRunner_Run_WritesSecrets(t *testing.T) {
	srv := newRunnerServer(t)
	defer srv.Close()

	client, err := vault.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	filter := sync.NewNamespaceFilter("app")
	syncer := sync.NewSyncer()
	writer := &stubWriter{}

	runner := sync.NewRunner(client, filter, syncer, writer, "secret")
	if err := runner.Run(); err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}

	if len(writer.data) == 0 {
		t.Error("expected secrets to be written, got none")
	}
}

func TestRunner_Run_WriterError(t *testing.T) {
	srv := newRunnerServer(t)
	defer srv.Close()

	client, err := vault.NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	filter := sync.NewNamespaceFilter("app")
	syncer := sync.NewSyncer()
	writer := &stubWriter{err: errors.New("disk full")}

	runner := sync.NewRunner(client, filter, syncer, writer, "secret")
	if err := runner.Run(); err == nil {
		t.Error("expected error from writer, got nil")
	}
}
