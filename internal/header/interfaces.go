package header

import (
	"io"
)

type HeaderComponent interface {
	Size() int
	Validate([]byte) error
}

type HeaderSerializer interface {
	Serialize(any) ([]byte, error)
	Deserialize([]byte, any) error
}

type HeaderIO interface {
	WriteComponent(w io.Writer, component HeaderComponent) error
	ReadComponent(r io.Reader, size int) ([]byte, error)
}

const (
	SaltSize          = 32
	OriginalSizeBytes = 8
	AesNonceSize      = 12
	ChaCha20NonceSize = 24
)
