package processor

import (
	"fmt"

	"github.com/hambosto/go-encryption/internal/algorithms"
	"github.com/hambosto/go-encryption/internal/encoding"
)

type Processor struct {
	AesCipher      *algorithms.AESCipher
	ChaCha20Cipher *algorithms.ChaCha20Cipher
	Encoder        *encoding.ReedSolomon
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

	encoder, err := encoding.NewReedSolomon(4, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to create Reed-Solomon encoder: %w", err)
	}

	return &Processor{
		AesCipher:      aesCipher,
		ChaCha20Cipher: chaCha20Cipher,
		Encoder:        encoder,
		IsEncryption:   isEncryption,
	}, nil
}

func (p *Processor) ProcessChunk(chunk []byte) ([]byte, error) {
	if p.IsEncryption {
		return p.encrypt(chunk)
	}
	return p.decrypt(chunk)
}
