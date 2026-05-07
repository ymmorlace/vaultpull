package sync_test

import (
	"context"
	"testing"

	"github.com/user/vaultpull/internal/sync"
)

// TestEncryptor_IntegrationWithDedupWriter verifies that the Encryptor
// composes correctly with DedupWriter: duplicate keys are rejected even
// when their encrypted forms differ (since encryption is non-deterministic).
func TestEncryptor_IntegrationWithDedupWriter(t *testing.T) {
	cap := &captureWriter{}
	dedup := sync.NewDedupWriter(cap)
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 7)
	}
	enc, err := sync.NewEncryptor(dedup, key)
	if err != nil {
		t.Fatalf("NewEncryptor: %v", err)
	}
	ctx := context.Background()

	if err := enc.Write(ctx, "DB_PASSWORD", "s3cr3t"); err != nil {
		t.Fatalf("first write failed: %v", err)
	}
	if err := enc.Write(ctx, "DB_PASSWORD", "other"); err == nil {
		t.Fatal("expected duplicate key error, got nil")
	}
}

// TestEncryptor_IntegrationRoundTrip verifies that values written through
// the Encryptor can be recovered with Decrypt using the same key.
func TestEncryptor_IntegrationRoundTrip(t *testing.T) {
	cases := []struct {
		key   string
		value string
	}{
		{"API_KEY", "abc123"},
		{"EMPTY_VAL", ""},
		{"UNICODE", "こんにちは"},
		{"LONG", string(make([]byte, 512))},
	}

	encKey := make([]byte, 32)
	for i := range encKey {
		encKey[i] = byte(255 - i)
	}

	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			w := &captureWriter{}
			enc, _ := sync.NewEncryptor(w, encKey)
			_ = enc.Write(context.Background(), tc.key, tc.value)

			got, err := sync.Decrypt(encKey, w.value)
			if err != nil {
				t.Fatalf("Decrypt: %v", err)
			}
			if got != tc.value {
				t.Errorf("round-trip mismatch: want %q, got %q", tc.value, got)
			}
		})
	}
}
