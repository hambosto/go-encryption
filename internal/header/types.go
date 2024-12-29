package header

import "io"

type HeaderReader interface {
	Read(r io.Reader) (FileHeader, error)
}

type HeaderWriter interface {
	Write(w io.Writer, header FileHeader) error
}
