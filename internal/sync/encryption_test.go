package sync_test

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	"github.com/user/vaultpull/internal/sync"
)

func TestEncryptor_PanicOnNilInner(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nil inner writer")
		}
	}()
	key := make([]byte, 32)
	_, _ = sync.NewEncryptor(nil, key)
}

func TestEncryptor_RejectsShortKey(t *testing.T) {
	w := &captureWriter{}
	_, err := sync.NewEncryptor(w, []byte("tooshort"))
	if err == nil {
		t.Fatal("expected error for short key")
	}
}

func TestEncryptor_WriteEncryptsValue(t *testing.T) {
	w := &captureWriter{}
	key := make([]byte, 32)
	enc, err := sync.NewEncryptor(w, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx := context.Background()
	if err := enc.Write(ctx, "SECRET_KEY", "plaintext"); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if w.key != "SECRET_KEY" {
		t.Errorf("expected key SECRET_KEY, got %s", w.key)
	}
	// Value should be base64-encoded ciphertext, not the original
	if w.value == "plaintext" {
		t.Error("value should be encrypted, not plaintext")
	}
	if _, err := base64.StdEncoding.DecodeString(w.value); err != nil {
		t.Errorf("value is not valid base64: %v", err)
	}
}

func TestEncryptor_DecryptRoundTrip(t *testing.T) {
	w := &captureWriter{}
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	enc, _ := sync.NewEncryptor(w, key)
	original := "super-secret-value"
	_ = enc.Write(context.Background(), "K", original)

	decrypted, err := sync.Decrypt(key, w.value)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if decrypted != original {
		t.Errorf("expected %q, got %q", original, decrypted)
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	w := &captureWriter{}
	key := make([]byte, 32)
	enc, _ := sync.NewEncryptor(w, key)
	_ = enc.Write(context.Background(), "K", "value")

	wrongKey := make([]byte, 32)
	wrongKey[0] = 0xFF
	_, err := sync.Decrypt(wrongKey, w.value)
	if err == nil {
		t.Fatal("expected error when decrypting with wrong key")
	}
}

func TestDecrypt_InvalidBase64(t *testing.T) {
	key := make([]byte, 32)
	_, err := sync.Decrypt(key, "!!!not-base64!!!")
	if err == nil || !strings.Contains(err.Error(), "base64") {
		t.Errorf("expected base64 error, got %v", err)
	}
}

func TestEncryptor_PropagatesInnerError(t *testing.T) {
	sentinel := errors.New("inner write failed")
	w := &captureWriter{err: sentinel}
	key := make([]byte, 32)
	enc, _ := sync.NewEncryptor(w, key)
	err := enc.Write(context.Background(), "K", "v")
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

// captureWriter records the last key/value written.
type captureWriter struct {
	key   string
	value string
	err   error
}

func (c *captureWriter) Write(_ context.Context, key, value string) error {
	c.key = key
	c.value = value
	return c.err
}
