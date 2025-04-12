package processor

import (
	"encoding/binary"
	"fmt"

	"github.com/hambosto/go-encryption/internal/compression"
)

func (c *ChunkProcessor) encrypt(chunk []byte) ([]byte, error) {
	compressedData, err := compression.CompressData(chunk)
	if err != nil {
		return nil, fmt.Errorf("Compression failed: %w", err)
	}

	sizeHeader := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeHeader, uint32(len(compressedData)))
	fullPayload := append(sizeHeader, compressedData...)

	// Pad to 16-byte boundary
	alignedSize := (len(fullPayload) + 15) & ^15
	paddedPayload := make([]byte, alignedSize)
	copy(paddedPayload, fullPayload)

	aesEncrypted, err := c.AESCipher.Encrypt(paddedPayload)
	if err != nil {
		return nil, fmt.Errorf("AES encryption failed: %w", err)
	}

	chaCha20Encrypted, err := c.ChaCha20Cipher.Encrypt(aesEncrypted)
	if err != nil {
		return nil, fmt.Errorf("ChaCha20 encryption failed: %w", err)
	}

	encoded, err := c.ReedSolomon.Encode(chaCha20Encrypted)
	if err != nil {
		return nil, fmt.Errorf("Reed-Solomon encoding failed: %w", err)
	}

	return encoded, nil
}
