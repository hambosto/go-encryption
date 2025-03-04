package header

import (
	"encoding/binary"
	"fmt"
	"io"
)

type BinaryHeaderIO struct{}

func NewBinaryHeaderIO() HeaderIO {
	return &BinaryHeaderIO{}
}

func (bio *BinaryHeaderIO) WriteComponent(w io.Writer, component HeaderComponent) error {
	switch c := component.(type) {
	case Salt:
		return bio.write(w, c.Value)
	case OriginalSize:
		buf := make([]byte, OriginalSizeBytes)
		binary.BigEndian.PutUint64(buf, c.Value)
		return bio.write(w, buf)
	case AesNonce:
		return bio.write(w, c.Value)
	case ChaCha20Nonce:
		return bio.write(w, c.Value)
	default:
		return fmt.Errorf("unsupported component type")
	}
}

func (bio *BinaryHeaderIO) ReadComponent(r io.Reader, size int) ([]byte, error) {
	buf := make([]byte, size)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, fmt.Errorf("reading component: %w", err)
	}
	return buf, nil
}

func (bio *BinaryHeaderIO) write(w io.Writer, data []byte) error {
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("writing component: %w", err)
	}
	return nil
}
