package encoding

import (
	"encoding/binary"
	"fmt"
)

func (r *ReedSolomonEncoder) Decode(data []byte) ([]byte, error) {
	if err := validateEncodedData(data, r.dataShards+r.parityShards); err != nil {
		return nil, err
	}

	decodedData, err := r.reconstructAndDecode(data)
	if err != nil {
		return nil, fmt.Errorf("decoding failed: %w", err)
	}

	return decodedData, nil
}

func validateEncodedData(data []byte, totalShards int) error {
	if len(data) == 0 || len(data)%totalShards != 0 {
		return NewValidationError("invalid encoded data size")
	}
	return nil
}

func (r *ReedSolomonEncoder) reconstructAndDecode(data []byte) ([]byte, error) {
	shards := r.splitIntoDecodingShards(data)

	if err := r.encoder.Reconstruct(shards); err != nil {
		return nil, fmt.Errorf("reconstruction failed: %w", err)
	}

	return r.extractOriginalData(shards)
}

func (r *ReedSolomonEncoder) splitIntoDecodingShards(data []byte) [][]byte {
	totalShards := r.dataShards + r.parityShards
	shardSize := len(data) / totalShards
	shards := make([][]byte, totalShards)

	for i := range shards {
		shards[i] = data[i*shardSize : (i+1)*shardSize]
	}

	return shards
}

func (r *ReedSolomonEncoder) extractOriginalData(shards [][]byte) ([]byte, error) {
	shardSize := len(shards[0])
	result := make([]byte, shardSize*r.dataShards)

	for i := range r.dataShards {
		copy(result[i*shardSize:], shards[i])
	}

	if len(result) < headerLength {
		return nil, NewValidationError("corrupted data: too short")
	}

	originalSize := binary.BigEndian.Uint32(result[:headerLength])
	if originalSize > uint32(len(result)-headerLength) {
		return nil, NewValidationError("corrupted data: invalid size header")
	}

	return result[headerLength : headerLength+originalSize], nil
}
