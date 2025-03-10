package cipher

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
	validKeySizes := map[int]bool{
		16: true,
		24: true,
		32: true,
	}

	if !validKeySizes[len(key)] {
		return nil, fmt.Errorf("AES key must be 16, 24, or 32 bytes")
	}

	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	return &AESCipher{
		key:   key,
		nonce: nonce,
	}, nil
}

func (c *AESCipher) Encrypt(plaintext []byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return nil, fmt.Errorf("plaintext cannot be empty")
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES GCM: %w", err)
	}

	ciphertext := aead.Seal(nil, c.nonce, plaintext, nil)
	return ciphertext, nil
}

func (c *AESCipher) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) == 0 {
		return nil, fmt.Errorf("ciphertext cannot be empty")
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES GCM: %w", err)
	}

	plaintext, err := aead.Open(nil, c.nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt ciphertext: %w", err)
	}

	return plaintext, nil
}

func (c *AESCipher) GetNonce() []byte {
	return c.nonce
}

func (c *AESCipher) SetNonce(nonce []byte) error {
	if len(nonce) != 12 {
		return fmt.Errorf("invalid nonce size: %d bytes", len(nonce))
	}
	c.nonce = nonce
	return nil
}
