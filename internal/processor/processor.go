package processor

import (
	"fmt"

	"github.com/hambosto/go-encryption/internal/algorithms"
	"github.com/hambosto/go-encryption/internal/encoding"
)

type Processor struct {
	aesCipher      *algorithms.AESCipher
	chaCha20Cipher *algorithms.ChaCha20Cipher
	encoder        *encoding.Encoder
	isEncryption   bool
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

	encoder, err := encoding.New(4, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to create Reed-Solomon encoder: %w", err)
	}

	return &Processor{
		aesCipher:      aesCipher,
		chaCha20Cipher: chaCha20Cipher,
		encoder:        encoder,
		isEncryption:   isEncryption,
	}, nil
}

func (p *Processor) ProcessChunk(chunk []byte) ([]byte, error) {
	if p.isEncryption {
		return p.encrypt(chunk)
	}
	return p.decrypt(chunk)
}
