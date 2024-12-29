package header

import (
	"encoding/binary"
	"fmt"
	"io"
)

type BinaryHeaderReader struct{}

func NewBinaryHeaderReader() *BinaryHeaderReader {
	return &BinaryHeaderReader{}
}

func (r *BinaryHeaderReader) Read(reader io.Reader) (FileHeader, error) {
	builder := NewFileHeaderBuilder()

	salt := make([]byte, 32)
	if _, err := io.ReadFull(reader, salt); err != nil {
		return FileHeader{}, fmt.Errorf("failed to read salt: %w", err)
	}
	builder.SetSalt(salt)

	sizeBytes := make([]byte, 8)
	if _, err := io.ReadFull(reader, sizeBytes); err != nil {
		return FileHeader{}, fmt.Errorf("failed to read original size: %w", err)
	}
	builder.SetOriginalSize(binary.BigEndian.Uint64(sizeBytes))

	aesNonce := make([]byte, 12)
	if _, err := io.ReadFull(reader, aesNonce); err != nil {
		return FileHeader{}, fmt.Errorf("failed to read aes nonce: %w", err)
	}
	builder.SetAesNonce(aesNonce)

	chaCha20Nonce := make([]byte, 24)
	if _, err := io.ReadFull(reader, chaCha20Nonce); err != nil {
		return FileHeader{}, fmt.Errorf("failed to read chacha20 nonce: %w", err)
	}
	builder.SetChaCha20Nonce(chaCha20Nonce)

	return builder.Build(), nil
}
