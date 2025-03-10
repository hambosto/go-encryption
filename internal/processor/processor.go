package processor

import (
	"fmt"

	"github.com/hambosto/go-encryption/internal/algorithms"
	"github.com/hambosto/go-encryption/internal/encoding"
)

type Processor struct {
	AESCipher      *algorithms.AESCipher
	ChaCha20Cipher *algorithms.ChaCha20Cipher
	ReedSolomon    *encoding.ReedSolomonEncoder
	IsEncryption   bool
}

func NewProcessor(key []byte, isEncryption bool) (*Processor, error) {
	if len(key) < 64 {
		return nil, fmt.Errorf("encryption key must be at least 64 bytes long")
	}

	aesCipher, err := algorithms.NewAESCipher(key[:32])
	if err != nil {
		return nil, fmt.Errorf("failed to create aes cipher: %w", err)
	}

	chaCha20Cipher, err := algorithms.NewChaCha20Cipher(key[32:64])
	if err != nil {
		return nil, fmt.Errorf("failed to create ChaCha20 cipher: %w", err)
	}

	reedSolomon, err := encoding.NewReedSolomonEncoder(encoding.Config{DataShards: 4, ParityShards: 10})
	if err != nil {
		return nil, fmt.Errorf("failed to create Reed-Solomon encoder: %w", err)
	}

	return &Processor{
		AESCipher:      aesCipher,
		ChaCha20Cipher: chaCha20Cipher,
		ReedSolomon:    reedSolomon,
		IsEncryption:   isEncryption,
	}, nil
}

func (p *Processor) ProcessChunk(chunk []byte) ([]byte, error) {
	if p.IsEncryption {
		return p.encrypt(chunk)
	}
	return p.decrypt(chunk)
}
