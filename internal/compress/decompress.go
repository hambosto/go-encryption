package compress

import (
	"encoding/binary"
	"fmt"

	"github.com/pierrec/lz4/v4"
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
	initialSize := len(compressedData) * 4
	decompressed := make([]byte, initialSize)

	n, err := lz4.UncompressBlock(compressedData, decompressed)
	if err == lz4.ErrInvalidSourceShortBuffer {
		decompressed = make([]byte, initialSize*2)
		n, err = lz4.UncompressBlock(compressedData, decompressed)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to decompress data with lz4: %w", err)
	}

	return decompressed[:n], nil
}
