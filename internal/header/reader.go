package header

import (
	"encoding/binary"
	"io"
)

type HeaderReader struct {
	io HeaderIO
}

func NewHeaderReader(io HeaderIO) *HeaderReader {
	return &HeaderReader{io: io}
}

func (r *HeaderReader) Read(reader io.Reader) (Header, error) {
	builder := NewHeaderBuilder()

	saltData, err := r.io.ReadComponent(reader, SaltSize)
	if err != nil {
		return Header{}, err
	}

	sizeData, err := r.io.ReadComponent(reader, OriginalSizeBytes)
	if err != nil {
		return Header{}, err
	}

	aesNonce, err := r.io.ReadComponent(reader, AesNonceSize)
	if err != nil {
		return Header{}, err
	}

	chaCha20Nonce, err := r.io.ReadComponent(reader, ChaCha20NonceSize)
	if err != nil {
		return Header{}, err
	}

	return builder.
		WithSalt(saltData).
		WithOriginalSize(binary.BigEndian.Uint64(sizeData)).
		WithAesNonce(aesNonce).
		WithChaCha20Nonce(chaCha20Nonce).
		Build()
}
