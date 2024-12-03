package encoding

import (
	"encoding/binary"
	"fmt"

	"github.com/vivint/infectious"
)

func reconstructShares(data []byte, totalShards int) ([]infectious.Share, error) {
	shareSize := len(data) / totalShards
	shares := make([]infectious.Share, totalShards)

	for i := 0; i < totalShards; i++ {
		start := i * shareSize
		shares[i] = infectious.Share{
			Data:   data[start : start+shareSize],
			Number: i,
		}
	}

	return shares, nil
}

func extractOriginalData(decoded []byte) ([]byte, error) {
	if len(decoded) < 4 {
		return nil, fmt.Errorf("decoded data too short: %d bytes", len(decoded))
	}

	originalSize := binary.BigEndian.Uint32(decoded[:4])
	if originalSize > uint32(len(decoded)-4) {
		return nil, fmt.Errorf("invalid size prefix: %d bytes", len(decoded))
	}

	return decoded[4 : 4+originalSize], nil
}
