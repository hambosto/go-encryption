package algorithms

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
)

const (
	aesNonceSize = 12
)

type aesCipher struct {
	key   []byte
	nonce []byte
}

func NewAESCipher(key []byte) (Cipher, error) {
	validKeySizes := map[int]bool{16: true, 24: true, 32: true}
	if !validKeySizes[len(key)] {
		return nil, fmt.Errorf("%w: AES key must be 16, 24, or 32 bytes", ErrInvalidKeySize)
	}

	nonce := make([]byte, aesNonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNonceGenerationFailed, err)
	}

	return &aesCipher{
		key:   key,
		nonce: nonce,
	}, nil
}

func (c *aesCipher) Encrypt(plaintext []byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return nil, ErrEmptyPlaintext
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCipherCreationFailed, err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCipherCreationFailed, err)
	}

	return aead.Seal(nil, c.nonce, plaintext, nil), nil
}

func (c *aesCipher) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCipherCreationFailed, err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCipherCreationFailed, err)
	}

	plaintext, err := aead.Open(nil, c.nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	return plaintext, nil
}

func (c *aesCipher) SetNonce(nonce []byte) error {
	if len(nonce) != aesNonceSize {
		return fmt.Errorf("%w: expected %d bytes, got %d", ErrInvalidNonceSize, aesNonceSize, len(nonce))
	}
	c.nonce = nonce
	return nil
}

func (c *aesCipher) GetNonce() []byte {
	return c.nonce
}
