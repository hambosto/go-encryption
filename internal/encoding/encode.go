package encoding

import (
	"encoding/binary"
	"fmt"
)

func (r *ReedSolomonEncoder) Encode(data []byte) ([]byte, error) {
	if err := validateData(data); err != nil {
		return nil, err
	}

	encodedData, err := r.prepareAndEncode(data)
	if err != nil {
		return nil, fmt.Errorf("encoding failed: %w", err)
	}

	return encodedData, nil
}

func validateData(data []byte) error {
	if len(data) == 0 || len(data) > maxDataSize {
		return NewValidationError(fmt.Sprintf("invalid data size: must be between 1 and %d bytes", maxDataSize))
	}
	return nil
}

func (r *ReedSolomonEncoder) prepareAndEncode(data []byte) ([]byte, error) {
	dataWithHeader := addHeader(data)
	shards := r.splitIntoShards(dataWithHeader)

	if err := r.encoder.Encode(shards); err != nil {
		return nil, err
	}

	return r.joinShards(shards), nil
}

func addHeader(data []byte) []byte {
	dataWithHeader := make([]byte, headerLength+len(data))
	binary.BigEndian.PutUint32(dataWithHeader, uint32(len(data)))
	copy(dataWithHeader[headerLength:], data)
	return dataWithHeader
}

func (r *ReedSolomonEncoder) splitIntoShards(data []byte) [][]byte {
	totalShards := r.dataShards + r.parityShards
	shardSize := (len(data) + r.dataShards - 1) / r.dataShards

	if shardSize%r.dataShards != 0 {
		shardSize = ((shardSize + r.dataShards - 1) / r.dataShards) * r.dataShards
	}

	shards := make([][]byte, totalShards)
	for i := range shards {
		shards[i] = make([]byte, shardSize)
	}

	for i := range data {
		shardNum := i / shardSize
		indexInShard := i % shardSize
		shards[shardNum][indexInShard] = data[i]
	}

	return shards
}

func (r *ReedSolomonEncoder) joinShards(shards [][]byte) []byte {
	shardSize := len(shards[0])
	result := make([]byte, shardSize*(r.dataShards+r.parityShards))

	for i, shard := range shards {
		copy(result[i*shardSize:], shard)
	}

	return result
}
