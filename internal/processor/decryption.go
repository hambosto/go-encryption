package processor

import (
	"fmt"

	"github.com/hambosto/go-encryption/internal/compression"
)

func (p *Processor) decrypt(chunk []byte) ([]byte, error) {
	decodedData, err := p.ReedSolomon.Decode(chunk)
	if err != nil {
		return nil, fmt.Errorf("reed-solomon decoding failed: %w", err)
	}

	chaCha20Decrypted, err := p.ChaCha20Cipher.Decrypt(decodedData)
	if err != nil {
		return nil, fmt.Errorf("ChaCha20 decryption failed: %w", err)
	}

	aesDecrypted, err := p.AESCipher.Decrypt(chaCha20Decrypted)
	if err != nil {
		return nil, fmt.Errorf("aes decryption failed: %w", err)
	}

	zlibDecompressed, err := compression.DecompressData(aesDecrypted)
	if err != nil {
		return nil, fmt.Errorf("zlib decompression failed: %w", err)
	}

	return zlibDecompressed, nil
}
