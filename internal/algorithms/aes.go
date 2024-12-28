package algorithms

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
)

type AESCipher struct {
	key   []byte
	nonce []byte
}

func NewAESCipher(key []byte) (*AESCipher, error) {
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	return &AESCipher{
		key:   key,
		nonce: nonce,
	}, nil
}

func (c *AESCipher) Encrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create AEAD: %w", err)
	}

	return aead.Seal(nil, c.nonce, ciphertext, nil), nil
}

func (c *AESCipher) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create AEAD: %v", err)
	}

	return aead.Seal(nil, c.nonce, ciphertext, nil), nil
}

func (c *AESCipher) SetNonce(nonce []byte) error {
	if len(nonce) != 12 {
		return fmt.Errorf("invalid nonce size: %d bytes", len(nonce))
	}

	c.nonce = nonce
	return nil
}

func (c *AESCipher) GetNonce() []byte {
	return c.nonce
}
