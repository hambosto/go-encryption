package header

import (
	"io"
)

type HeaderWriter struct {
	io HeaderIO
}

func NewHeaderWriter(io HeaderIO) *HeaderWriter {
	return &HeaderWriter{io: io}
}

func (w *HeaderWriter) Write(writer io.Writer, header Header) error {
	components := []HeaderComponent{
		header.Salt,
		header.OriginalSize,
		header.AesNonce,
		header.ChaCha20Nonce,
	}

	for _, component := range components {
		if err := w.io.WriteComponent(writer, component); err != nil {
			return err
		}
	}

	return nil
}
