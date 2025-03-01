package processor

import (
	"fmt"

	"github.com/hambosto/go-encryption/internal/algorithms"
	"github.com/hambosto/go-encryption/internal/encoding"
)

type Processor struct {
	primaryCipher   algorithms.Cipher
	secondaryCipher algorithms.Cipher
	reedsolomon     *encoding.ReedSolomonEncoder
	IsEncryption    bool
}

func NewProcessor(key []byte, isEncryption bool) (*Processor, error) {
	if len(key) < 64 {
		return nil, fmt.Errorf("encryption key must be at least 64 bytes long")
	}

	primaryCipher, err := algorithms.NewCipher(algorithms.AES, key[:32])
	if err != nil {
		return nil, fmt.Errorf("failed to create aes cipher: %w", err)
	}

	secondaryCipher, err := algorithms.NewCipher(algorithms.CHACHA20, key[32:64])
	if err != nil {
		return nil, fmt.Errorf("failed to create ChaCha20 cipher: %w", err)
	}

	encoder, err := encoding.NewReedSolomonEncoder(encoding.Config{DataShards: 4, ParityShards: 10})
	if err != nil {
		return nil, fmt.Errorf("failed to create Reed-Solomon encoder: %w", err)
	}

	return &Processor{
		primaryCipher:   primaryCipher,
		secondaryCipher: secondaryCipher,
		reedsolomon:     encoder,
		IsEncryption:    isEncryption,
	}, nil
}

func (p *Processor) ProcessChunk(chunk []byte) ([]byte, error) {
	if p.IsEncryption {
		return p.encrypt(chunk)
	}
	return p.decrypt(chunk)
}

func (p *Processor) GetPrimaryCipher() algorithms.Cipher {
	return p.primaryCipher
}

func (p *Processor) GetSecondaryCipher() algorithms.Cipher {
	return p.secondaryCipher
}
