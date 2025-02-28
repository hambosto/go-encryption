package algorithms

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
)

const (
	chaCha20NonceSize = 24
)

type chaCha20Cipher struct {
	key   []byte
	nonce []byte
}

func NewChaCha20Cipher(key []byte) (Cipher, error) {
	if len(key) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("%w: expected %d bytes, got %d", ErrInvalidKeySize, chacha20poly1305.KeySize, len(key))
	}

	nonce := make([]byte, chaCha20NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNonceGenerationFailed, err)
	}

	return &chaCha20Cipher{
		key:   key,
		nonce: nonce,
	}, nil
}

func (c *chaCha20Cipher) Encrypt(plaintext []byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return nil, ErrEmptyPlaintext
	}

	aead, err := chacha20poly1305.New(c.key)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCipherCreationFailed, err)
	}

	nonce := c.nonce[:chacha20poly1305.NonceSize]
	return aead.Seal(nil, nonce, plaintext, nil), nil
}

func (c *chaCha20Cipher) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < chacha20poly1305.Overhead {
		return nil, fmt.Errorf("%w: need at least %d bytes", ErrCiphertextTooShort, chacha20poly1305.Overhead)
	}

	aead, err := chacha20poly1305.New(c.key)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCipherCreationFailed, err)
	}

	nonce := c.nonce[:chacha20poly1305.NonceSize]
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	return plaintext, nil
}

func (c *chaCha20Cipher) SetNonce(nonce []byte) error {
	if len(nonce) != chaCha20NonceSize {
		return fmt.Errorf("%w: expected %d bytes, got %d", ErrInvalidNonceSize, chaCha20NonceSize, len(nonce))
	}
	c.nonce = nonce
	return nil
}

func (c *chaCha20Cipher) GetNonce() []byte {
	return c.nonce
}
