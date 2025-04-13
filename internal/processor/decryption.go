package processor

import (
	"fmt"

	"github.com/hambosto/go-encryption/internal/compression"
)

func (c *ChunkProcessor) decrypt(chunk []byte) ([]byte, error) {
	decodedData, err := c.ReedSolomon.Decode(chunk)
	if err != nil {
		return nil, fmt.Errorf("reed-solomon decoding failed: %w", err)
	}

	chaCha20Decrypted, err := c.ChaCha20Cipher.Decrypt(decodedData)
	if err != nil {
		return nil, fmt.Errorf("ChaCha20 decryption failed: %w", err)
	}

	aesDecrypted, err := c.AESCipher.Decrypt(chaCha20Decrypted)
	if err != nil {
		return nil, fmt.Errorf("AES decryption failed: %w", err)
	}

	decompressedData, err := compression.DecompressData(aesDecrypted)
	if err != nil {
		return nil, fmt.Errorf("zlib decompression failed: %w", err)
	}

	return decompressedData, nil
}
