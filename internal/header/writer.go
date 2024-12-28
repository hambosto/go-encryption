package header

import (
	"encoding/binary"
	"fmt"
	"io"
)

func Write(w io.Writer, header FileHeader) error {
	if _, err := w.Write(header.Salt); err != nil {
		return fmt.Errorf("failed to write salt: %w", err)
	}

	sizeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(sizeBytes, header.OriginalSize)
	if _, err := w.Write(sizeBytes); err != nil {
		return fmt.Errorf("failed to write original size: %w", err)
	}

	if _, err := w.Write(header.AesNonce); err != nil {
		return fmt.Errorf("failed to write aes nonce: %w", err)
	}

	if _, err := w.Write(header.ChaCha20Nonce); err != nil {
		return fmt.Errorf("failed to write chacha20 nonce: %w", err)
	}

	return nil
}
