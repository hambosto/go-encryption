package processor

import (
	"fmt"

	"github.com/hambosto/go-encryption/internal/cipher"
	"github.com/hambosto/go-encryption/internal/encoding"
)

type ChunkProcessor struct {
	AESCipher      *cipher.AESCipher
	ChaCha20Cipher *cipher.ChaCha20Cipher
	ReedSolomon    *encoding.ReedSolomon
	IsEncryption   bool
}

func NewChunkProcessor(key []byte, isEncryption bool) (*ChunkProcessor, error) {
	if len(key) < 64 {
		return nil, fmt.Errorf("encryption key must be at least 64 bytes long")
	}

	aesCipher, err := cipher.NewAESCipher(key[:32])
	if err != nil {
		return nil, fmt.Errorf("failed to create aes cipher: %w", err)
	}

	chaCha20Cipher, err := cipher.NewChaCha20Cipher(key[32:64])
	if err != nil {
		return nil, fmt.Errorf("failed to create ChaCha20 cipher: %w", err)
	}

	reedSolomon, err := encoding.NewReedSolomon(encoding.ReedSolomonConfig{DataShards: 4, ParityShards: 10})
	if err != nil {
		return nil, fmt.Errorf("failed to create Reed-Solomon encoder: %w", err)
	}

	return &ChunkProcessor{
		AESCipher:      aesCipher,
		ChaCha20Cipher: chaCha20Cipher,
		ReedSolomon:    reedSolomon,
		IsEncryption:   isEncryption,
	}, nil
}

func (c *ChunkProcessor) ProcessChunk(chunk []byte) ([]byte, error) {
	if c.IsEncryption {
		return c.encrypt(chunk)
	}
	return c.decrypt(chunk)
}
