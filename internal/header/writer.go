package header

import (
	"encoding/binary"
	"fmt"
	"io"
)

type BinaryHeaderWriter struct{}

func NewBinaryHeaderWriter() *BinaryHeaderWriter {
	return &BinaryHeaderWriter{}
}

func (w *BinaryHeaderWriter) Write(writer io.Writer, header FileHeader) error {
	if _, err := writer.Write(header.Salt); err != nil {
		return fmt.Errorf("failed to write salt: %w", err)
	}

	sizeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(sizeBytes, header.OriginalSize)
	if _, err := writer.Write(sizeBytes); err != nil {
		return fmt.Errorf("failed to write original size: %w", err)
	}

	if _, err := writer.Write(header.AesNonce); err != nil {
		return fmt.Errorf("failed to write aes nonce: %w", err)
	}

	if _, err := writer.Write(header.ChaCha20Nonce); err != nil {
		return fmt.Errorf("failed to write chacha20 nonce: %w", err)
	}

	return nil
}
