package header

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/hambosto/go-encryption/pkg/crypto/config"
)

type FileHeader struct {
	Salt          []byte
	OriginalSize  uint64
	SerpentNonce  []byte
	ChaCha20Nonce []byte
}

func Write(w io.Writer, header FileHeader) error {
	if _, err := w.Write(header.Salt); err != nil {
		return fmt.Errorf("failed to write salt: %w", err)
	}

	sizeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(sizeBytes, header.OriginalSize)

	if _, err := w.Write(sizeBytes); err != nil {
		return fmt.Errorf("failed to write original size: %w", err)
	}

	if _, err := w.Write(header.SerpentNonce); err != nil {
		return fmt.Errorf("failed to write serpent nonce: %w", err)
	}

	if _, err := w.Write(header.ChaCha20Nonce); err != nil {
		return fmt.Errorf("failed to write chacha20 nonce: %w", err)
	}

	return nil
}

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
