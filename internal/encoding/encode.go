package encoding

import (
	"encoding/binary"
	"fmt"
)

func (e *Encoder) Encode(data []byte) ([]byte, error) {
	if len(data) == 0 || len(data) > maxDataSize {
		return nil, fmt.Errorf("invalid data size: must be between 1 and %d bytes", maxDataSize)
	}

	dataWithHeader := make([]byte, headerLength+len(data))
	binary.BigEndian.PutUint32(dataWithHeader, uint32(len(data)))
	copy(dataWithHeader[headerLength:], data)

	shards, size := e.splitIntoShards(dataWithHeader)

	if err := e.encoder.Encode(shards); err != nil {
		return nil, fmt.Errorf("encoding failed: %w", err)
	}

	return e.joinShards(shards, size), nil
}
