package processor

import (
	"fmt"
)

func (p *Processor) encrypt(chunk []byte) ([]byte, error) {
	compressedData, err := p.compressData(chunk)
	if err != nil {
		return nil, fmt.Errorf("compression failed: %w", err)
	}

	paddedData := p.padData(compressedData)

	aesEncrypted, err := p.AesCipher.Encrypt(paddedData)
	if err != nil {
		return nil, fmt.Errorf("aes encryption failed: %w", err)
	}

	chaCha20Encrypted, err := p.ChaCha20Cipher.Encrypt(aesEncrypted)
	if err != nil {
		return nil, fmt.Errorf("ChaCha20 encryption failed: %w", err)
	}

	encoded, err := p.Encoder.Encode(chaCha20Encrypted)
	if err != nil {
		return nil, fmt.Errorf("Reed-Solomon encoding failed: %w", err)
	}

	return encoded, nil
}
