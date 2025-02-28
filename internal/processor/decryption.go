package processor

import (
	"fmt"
)

func (p *Processor) decrypt(chunk []byte) ([]byte, error) {
	decodedData, err := p.reedsolomon.Decode(chunk)
	if err != nil {
		return nil, fmt.Errorf("reed-solomon decoding failed: %w", err)
	}

	chaCha20Decrypted, err := p.secondaryCipher.Decrypt(decodedData)
	if err != nil {
		return nil, fmt.Errorf("ChaCha20 decryption failed: %w", err)
	}

	aesDecrypted, err := p.primaryCipher.Decrypt(chaCha20Decrypted)
	if err != nil {
		return nil, fmt.Errorf("aes decryption failed: %w", err)
	}

	zlibDecompressed, err := p.decompressData(aesDecrypted)
	if err != nil {
		return nil, fmt.Errorf("zlib decompression failed: %w", err)
	}

	return zlibDecompressed, nil
}
