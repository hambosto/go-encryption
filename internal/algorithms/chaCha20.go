package algorithms

import (
	"crypto/rand"
	"fmt"

	"github.com/hambosto/go-encryption/internal/constants"
	"golang.org/x/crypto/chacha20poly1305"
)

type ChaCha20Cipher struct {
	key      []byte
	nonce    []byte
	baseNone []byte
}

func NewChaCha20Cipher(key []byte) (*ChaCha20Cipher, error) {
	nonce := make([]byte, constants.NonceSizeX)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	return &ChaCha20Cipher{
		key:   key,
		nonce: nonce,
	}, nil
}

func (c *ChaCha20Cipher) Encrypt(plaintext []byte) ([]byte, error) {
	if len(plaintext)%16 != 0 {
		padding := make([]byte, 16-(len(plaintext)%16))
		plaintext = append(plaintext, padding...)
	}

	aead, err := chacha20poly1305.New(c.key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, constants.NonceSize)
	copy(nonce, c.baseNone)

	return aead.Seal(nil, nonce, plaintext, nil), nil
}

func (c *ChaCha20Cipher) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < chacha20poly1305.Overhead {
		return nil, fmt.Errorf("ciphertext too short: %d bytes", len(ciphertext))
	}

	aead, err := chacha20poly1305.New(c.key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, constants.NonceSize)
	copy(nonce, c.baseNone)

	return aead.Open(nil, nonce, ciphertext, nil)
}

func (c *ChaCha20Cipher) SetNonce(nonce []byte) error {
	if len(nonce) != constants.NonceSizeX {
		return fmt.Errorf("invalid nonce size: %d bytes", len(nonce))
	}

	c.nonce = nonce
	return nil
}

func (c *ChaCha20Cipher) GetNonce() []byte {
	return c.nonce
}
