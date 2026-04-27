package vault_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultpull/internal/vault"
)

func TestReadSecret_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") != "test-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"data": map[string]interface{}{
					"DB_HOST": "localhost",
					"DB_PORT": "5432",
				},
			},
		})
	}))
	defer ts.Close()

	c, err := vault.NewClient(ts.URL, "test-token", "")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	data, err := c.ReadSecret("myapp/config")
	if err != nil {
		t.Fatalf("ReadSecret: %v", err)
	}
	if data["DB_HOST"] != "localhost" {
		t.Errorf("expected DB_HOST=localhost, got %q", data["DB_HOST"])
	}
	if data["DB_PORT"] != "5432" {
		t.Errorf("expected DB_PORT=5432, got %q", data["DB_PORT"])
	}
}

func TestReadSecret_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	c, err := vault.NewClient(ts.URL, "tok", "")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = c.ReadSecret("missing/path")
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
}

func TestReadSecret_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer ts.Close()

	c, err := vault.NewClient(ts.URL, "tok", "")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = c.ReadSecret("some/path")
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
}
