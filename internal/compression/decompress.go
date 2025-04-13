package compression

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
)

func DecompressData(data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("invalid data: insufficient bytes for size header")
	}

	compressedSize := binary.BigEndian.Uint32(data[:4])
	if compressedSize > uint32(len(data)) {
		return nil, fmt.Errorf("invalid compressed data size: expected %d, got %d bytes available",
			compressedSize, len(data)-4)
	}

	compressedData := data[4 : 4+compressedSize]

	return decompress(compressedData)
}

func decompress(compressedData []byte) ([]byte, error) {
	// Create a zlib reader for the compressed data
	r, err := zlib.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader: %w", err)
	}
	defer r.Close()

	// Read all the decompressed data
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return nil, fmt.Errorf("failed to decompress data with zlib: %w", err)
	}

	return buf.Bytes(), nil
}
