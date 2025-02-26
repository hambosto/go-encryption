package encoding

import (
	"encoding/binary"
	"fmt"
)

func (r *ReedSolomon) Decode(data []byte) ([]byte, error) {
	totalShards := r.dataShards + r.parityShards

	if len(data) == 0 || len(data)%totalShards != 0 {
		return nil, fmt.Errorf("invalid encoded data size")
	}

	shardSize := len(data) / totalShards
	shards := make([][]byte, totalShards)
	for i := range shards {
		shards[i] = data[i*shardSize : (i+1)*shardSize]
	}

	if err := r.encoder.Reconstruct(shards); err != nil {
		return nil, fmt.Errorf("reconstruction failed: %w", err)
	}

	dataSize := shardSize * r.dataShards
	result := make([]byte, dataSize)
	for i := range r.dataShards {
		copy(result[i*shardSize:], shards[i])
	}

	if len(result) < headerLength {
		return nil, fmt.Errorf("corrupted data: too short")
	}

	originalSize := binary.BigEndian.Uint32(result[:headerLength])
	if originalSize > uint32(len(result)-headerLength) {
		return nil, fmt.Errorf("corrupted data: invalid size header")
	}

	return result[headerLength : headerLength+originalSize], nil
}
