package header

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/hambosto/go-encryption/internal/config"
)

func Read(r io.Reader) (FileHeader, error) {
	header := FileHeader{
		Salt:          make([]byte, config.SaltSize),
		SerpentNonce:  make([]byte, config.NonceSize),
		ChaCha20Nonce: make([]byte, config.NonceSizeX),
	}

	if _, err := io.ReadFull(r, header.Salt); err != nil {
		return header, fmt.Errorf("failed to read salt: %w", err)
	}

	sizeBytes := make([]byte, 8)
	if _, err := io.ReadFull(r, sizeBytes); err != nil {
		return header, fmt.Errorf("failed to read original size: %w", err)
	}
	header.OriginalSize = binary.BigEndian.Uint64(sizeBytes)

	if _, err := io.ReadFull(r, header.SerpentNonce); err != nil {
		return header, fmt.Errorf("failed to read serpent nonce: %w", err)
	}

	if _, err := io.ReadFull(r, header.ChaCha20Nonce); err != nil {
		return header, fmt.Errorf("failed to read chacha20 nonce: %w", err)
	}

	return header, nil
}
