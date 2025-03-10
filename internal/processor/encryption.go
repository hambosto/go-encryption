package processor

import (
	"fmt"

	"github.com/hambosto/go-encryption/internal/compress"
	"github.com/hambosto/go-encryption/internal/padding"
)

func (p *Processor) encrypt(chunk []byte) ([]byte, error) {
	compressedData, err := compress.CompressData(chunk)
	if err != nil {
		return nil, fmt.Errorf("Compression failed: %w", err)
	}

	paddedData := padding.Pad(compressedData)

	aesEncrypted, err := p.AESCipher.Encrypt(paddedData)
	if err != nil {
		return nil, fmt.Errorf("AES encryption failed: %w", err)
	}

	chaCha20Encrypted, err := p.ChaCha20Cipher.Encrypt(aesEncrypted)
	if err != nil {
		return nil, fmt.Errorf("ChaCha20 encryption failed: %w", err)
	}

	encoded, err := p.ReedSolomon.Encode(chaCha20Encrypted)
	if err != nil {
		return nil, fmt.Errorf("Reed-Solomon encoding failed: %w", err)
	}

	return encoded, nil
}
