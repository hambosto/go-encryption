package compression

import (
	"fmt"

	"github.com/pierrec/lz4/v4"
)

func CompressData(data []byte) ([]byte, error) {
	maxCompressedSize := lz4.CompressBlockBound(len(data))
	compressed := make([]byte, maxCompressedSize)

	compressedSize, err := lz4.CompressBlock(data, compressed, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to compress data with lz4: %w", err)
	}

	return compressed[:compressedSize], nil
}
