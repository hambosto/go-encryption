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

	aesEncrypted, err := p.aesCipher.Encrypt(paddedData)
	if err != nil {
		return nil, fmt.Errorf("aes encryption failed: %w", err)
	}

	chaCha20Encrypted, err := p.chaCha20Cipher.Encrypt(aesEncrypted)
	if err != nil {
		return nil, fmt.Errorf("ChaCha20 encryption failed: %w", err)
	}

	encoded, err := p.encoder.Encode(chaCha20Encrypted)
	if err != nil {
		return nil, fmt.Errorf("Reed-Solomon encoding failed: %w", err)
	}

	return encoded, nil
}
