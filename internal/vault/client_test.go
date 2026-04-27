package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newMockVaultServer(t *testing.T, kvData map[string]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			data := make(map[string]interface{}, len(kvData))
			for k, v := range kvData {
				data[k] = v
			}
			resp := map[string]interface{}{
				"data": map[string]interface{}{
					"data": data,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
}

func TestNewClient_InvalidAddr(t *testing.T) {
	_, err := NewClient("://bad-addr", "token", "")
	if err == nil {
		t.Fatal("expected error for invalid address, got nil")
	}
}

func TestNewClient_ValidParams(t *testing.T) {
	client, err := NewClient("http://127.0.0.1:8200", "root", "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestListSecrets_KVv2(t *testing.T) {
	expected := map[string]string{
		"DB_HOST": "localhost",
		"DB_PASS": "s3cr3t",
	}

	server := newMockVaultServer(t, expected)
	defer server.Close()

	client, err := NewClient(server.URL, "test-token", "")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	got, err := client.ListSecrets(context.Background(), "myapp/prod")
	if err != nil {
		t.Fatalf("ListSecrets error: %v", err)
	}

	for k, v := range expected {
		if got[k] != v {
			t.Errorf("key %q: want %q, got %q", k, v, got[k])
		}
	}
}

func TestListSecrets_WithPrefix(t *testing.T) {
	expected := map[string]string{"API_KEY": "abc123"}

	server := newMockVaultServer(t, expected)
	defer server.Close()

	client, err := NewClient(server.URL, "test-token", "kv")
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	got, err := client.ListSecrets(context.Background(), "myapp")
	if err != nil {
		t.Fatalf("ListSecrets error: %v", err)
	}

	if got["API_KEY"] != "abc123" {
		t.Errorf("expected API_KEY=abc123, got %q", got["API_KEY"])
	}
}
