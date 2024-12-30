package processor

import (
	"fmt"
)

func (p *Processor) decrypt(chunk []byte) ([]byte, error) {
	decodedData, err := p.encoder.Decode(chunk)
	if err != nil {
		return nil, fmt.Errorf("reed-solomon decoding failed: %w", err)
	}

	chaCha20Decrypted, err := p.chaCha20Cipher.Decrypt(decodedData)
	if err != nil {
		return nil, fmt.Errorf("ChaCha20 decryption failed: %w", err)
	}

	aesDecrypted, err := p.aesCipher.Decrypt(chaCha20Decrypted)
	if err != nil {
		return nil, fmt.Errorf("aes decryption failed: %w", err)
	}

	zlibDecompressed, err := p.decompressData(aesDecrypted)
	if err != nil {
		return nil, fmt.Errorf("zlib decompression failed: %w", err)
	}

	return zlibDecompressed, nil
}
