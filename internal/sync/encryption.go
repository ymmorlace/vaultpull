package sync

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

// Encryptor encrypts secret values using AES-256-GCM before writing.
type Encryptor struct {
	inner  EnvWriter
	gcm    cipher.AEAD
}

// EnvWriter is the common writer interface used throughout the sync package.
type EnvWriter interface {
	Write(ctx context.Context, key, value string) error
}

// NewEncryptor returns an Encryptor that wraps inner and encrypts each value
// with AES-256-GCM using the provided 32-byte key.
func NewEncryptor(inner EnvWriter, keyBytes []byte) (*Encryptor, error) {
	if inner == nil {
		panic("encryption: inner writer must not be nil")
	}
	if len(keyBytes) != 32 {
		return nil, fmt.Errorf("encryption: key must be 32 bytes, got %d", len(keyBytes))
	}
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("encryption: failed to create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("encryption: failed to create GCM: %w", err)
	}
	return &Encryptor{inner: inner, gcm: gcm}, nil
}

// Write encrypts value and forwards the base64-encoded ciphertext to inner.
func (e *Encryptor) Write(ctx context.Context, key, value string) error {
	nonce := make([]byte, e.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("encryption: nonce generation failed: %w", err)
	}
	ciphertext := e.gcm.Seal(nonce, nonce, []byte(value), nil)
	encoded := base64.StdEncoding.EncodeToString(ciphertext)
	return e.inner.Write(ctx, key, encoded)
}

// Decrypt reverses the encryption applied by Encryptor.Write.
func Decrypt(keyBytes []byte, encoded string) (string, error) {
	if len(keyBytes) != 32 {
		return "", fmt.Errorf("encryption: key must be 32 bytes, got %d", len(keyBytes))
	}
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("encryption: base64 decode failed: %w", err)
	}
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return "", errors.New("encryption: ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("encryption: decryption failed: %w", err)
	}
	return string(plaintext), nil
}
