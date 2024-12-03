package algorithms

import (
	"crypto/cipher"
	"crypto/rand"
	"fmt"

	"github.com/aead/serpent"
	"github.com/hambosto/go-encryption/internal/constants"
)

type SerpentCipher struct {
	key   []byte
	nonce []byte
}

func NewSerpentCipher(key []byte) (*SerpentCipher, error) {
	nonce := make([]byte, constants.NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %v", err)
	}

	return &SerpentCipher{
		key:   key,
		nonce: nonce,
	}, nil
}

func (c *SerpentCipher) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := serpent.NewCipher(c.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %v", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create AEAD: %v", err)
	}

	return aead.Seal(nil, c.nonce, plaintext, nil), nil
}

func (c *SerpentCipher) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := serpent.NewCipher(c.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %v", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create AEAD: %v", err)
	}

	return aead.Seal(nil, c.nonce, ciphertext, nil), nil
}

func (c *SerpentCipher) SetNonce(nonce []byte) error {
	if len(nonce) != constants.NonceSize {
		return fmt.Errorf("invalid nonce size: %d bytes", len(nonce))
	}

	c.nonce = nonce
	return nil
}

func (c *SerpentCipher) GetNonce() []byte {
	return c.nonce
}
