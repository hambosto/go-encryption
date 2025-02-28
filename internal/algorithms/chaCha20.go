package algorithms

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
)

type ChaCha20Cipher struct {
	key   []byte
	nonce []byte
}

func NewChaCha20Cipher(key []byte) (*ChaCha20Cipher, error) {
	if len(key) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: %d bytes, expected %d bytes", len(key), chacha20poly1305.KeySize)
	}

	nonce := make([]byte, 24)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate random nonce: %w", err)
	}

	return &ChaCha20Cipher{key: key, nonce: nonce}, nil
}

func (c *ChaCha20Cipher) Encrypt(plaintext []byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return nil, fmt.Errorf("plaintext cannot be empty")
	}

	aead, err := chacha20poly1305.New(c.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AEAD instance: %w", err)
	}

	nonce := c.nonce[:chacha20poly1305.NonceSize]
	return aead.Seal(nil, nonce, plaintext, nil), nil
}

func (c *ChaCha20Cipher) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < chacha20poly1305.Overhead {
		return nil, fmt.Errorf("ciphertext too short: %d bytes", len(ciphertext))
	}

	aead, err := chacha20poly1305.New(c.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AEAD instance: %w", err)
	}

	nonce := c.nonce[:chacha20poly1305.NonceSize]
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

func (c *ChaCha20Cipher) SetNonce(nonce []byte) error {
	if len(nonce) != 24 {
		return fmt.Errorf("invalid nonce size: %d bytes, expected %d bytes", len(nonce), 24)
	}
	c.nonce = nonce
	return nil
}

func (c *ChaCha20Cipher) GetNonce() []byte {
	return c.nonce
}
