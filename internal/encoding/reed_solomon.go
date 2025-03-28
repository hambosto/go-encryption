package encoding

import (
	"encoding/binary"
	"fmt"

	"github.com/klauspost/reedsolomon"
)

const (
	headerLength = 4
	maxDataSize  = 1 << 30
)

type ReedSolomon struct {
	dataShards   int
	parityShards int
	encoder      reedsolomon.Encoder
}

func NewReedSolomon(config ReedSolomonConfig) (*ReedSolomon, error) {
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	encoder, err := reedsolomon.New(config.DataShards, config.ParityShards)
	if err != nil {
		return nil, fmt.Errorf("failed to create reed-solomon encoder: %w", err)
	}

	return &ReedSolomon{
		dataShards:   config.DataShards,
		parityShards: config.ParityShards,
		encoder:      encoder,
	}, nil
}

func (r *ReedSolomon) Encode(data []byte) ([]byte, error) {
	if err := validateData(data); err != nil {
		return nil, err
	}
	return r.prepareAndEncode(data)
}

func (r *ReedSolomon) Decode(data []byte) ([]byte, error) {
	if err := validateEncodedData(data, r.dataShards+r.parityShards); err != nil {
		return nil, err
	}
	return r.reconstructAndDecode(data)
}

func (r *ReedSolomon) prepareAndEncode(data []byte) ([]byte, error) {
	dataWithHeader := addHeader(data)
	shards := r.splitIntoShards(dataWithHeader)

	if err := r.encoder.Encode(shards); err != nil {
		return nil, fmt.Errorf("encoding failed: %w", err)
	}

	return r.joinShards(shards), nil
}

func (r *ReedSolomon) reconstructAndDecode(data []byte) ([]byte, error) {
	shards := r.splitIntoDecodingShards(data)

	if err := r.encoder.Reconstruct(shards); err != nil {
		return nil, fmt.Errorf("recontruction failed: %w", err)
	}

	return r.extractOriginalData(shards)
}

func (r *ReedSolomon) splitIntoShards(data []byte) [][]byte {
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

func (r *ReedSolomon) splitIntoDecodingShards(data []byte) [][]byte {
	totalShards := r.dataShards + r.parityShards
	shardSize := len(data) / totalShards
	shards := make([][]byte, totalShards)

	for i := range shards {
		shards[i] = data[i*shardSize : (i+1)*shardSize]
	}

	return shards
}

func (r *ReedSolomon) joinShards(shards [][]byte) []byte {
	shardSize := len(shards[0])
	result := make([]byte, shardSize*(r.dataShards+r.parityShards))

	for i, shard := range shards {
		copy(result[i*shardSize:], shard)
	}

	return result
}

func (r *ReedSolomon) extractOriginalData(shards [][]byte) ([]byte, error) {
	shardsSize := len(shards[0])
	result := make([]byte, shardsSize*r.dataShards)

	for i := range r.dataShards {
		copy(result[i*shardsSize:], shards[i])
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
