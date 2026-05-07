package sync

import (
	"crypto/aes"
	"crypto/cipher"
)

// NewEncryptorExported exposes NewEncryptor for white-box tests.
func NewEncryptorExported(inner EnvWriter, key []byte) (*Encryptor, error) {
	return NewEncryptor(inner, key)
}

// newTestGCM creates a GCM instance for unit testing internals.
func newTestGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
