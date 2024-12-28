package encoding

import (
	"encoding/binary"
	"fmt"

	"github.com/klauspost/reedsolomon"
)

type ReedSolomonEncoder struct {
	enc         reedsolomon.Encoder
	dataShards  int
	totalShards int
}

func NewReedSolomonEncoder(dataShards, parityShards int) (*ReedSolomonEncoder, error) {
	if err := validateConfig(dataShards, parityShards); err != nil {
		return nil, err
	}

	totalShards := dataShards + parityShards
	enc, err := reedsolomon.New(dataShards, parityShards)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize encoder: %w", err)
	}

	return &ReedSolomonEncoder{
		enc:         enc,
		dataShards:  dataShards,
		totalShards: totalShards,
	}, nil
}

func (r *ReedSolomonEncoder) Encode(data []byte) ([]byte, error) {
	if err := validateDataSize(data); err != nil {
		return nil, err
	}

	// Prepare data with size prefix
	sizePrefix := make([]byte, 4)
	binary.BigEndian.PutUint32(sizePrefix, uint32(len(data)))
	paddedData := append(sizePrefix, data...)

	// Calculate the shard size
	shardSize := (len(paddedData) + r.dataShards - 1) / r.dataShards
	if shardSize == 0 {
		shardSize = 1
	}

	// Create shards
	shards := make([][]byte, r.totalShards)
	for i := range shards {
		shards[i] = make([]byte, shardSize)
	}

	// Split the data into shards
	for i := 0; i < len(paddedData); i++ {
		shardNum := i / shardSize
		if shardNum >= r.dataShards {
			break
		}
		byteNum := i % shardSize
		shards[shardNum][byteNum] = paddedData[i]
	}

	// Encode parity
	if err := r.enc.Encode(shards); err != nil {
		return nil, fmt.Errorf("failed to encode data: %w", err)
	}

	// Combine all shards
	result := make([]byte, shardSize*r.totalShards)
	for i, shard := range shards {
		copy(result[i*shardSize:], shard)
	}

	return result, nil
}

func (r *ReedSolomonEncoder) Decode(data []byte) ([]byte, error) {
	if err := validateEncodedData(data, r.totalShards); err != nil {
		return nil, err
	}

	// Split the data into shards
	shardSize := len(data) / r.totalShards
	shards := make([][]byte, r.totalShards)
	for i := range shards {
		shards[i] = make([]byte, shardSize)
		copy(shards[i], data[i*shardSize:(i+1)*shardSize])
	}

	// Reconstruct missing shards if necessary
	if err := r.enc.Reconstruct(shards); err != nil {
		return nil, fmt.Errorf("failed to reconstruct data: %w", err)
	}

	// Verify the reconstruction if possible
	if ok, err := r.enc.Verify(shards); err != nil {
		return nil, fmt.Errorf("failed to verify reconstruction: %w", err)
	} else if !ok {
		return nil, fmt.Errorf("reconstruction verification failed")
	}

	// Combine data shards
	dataLen := shardSize * r.dataShards
	result := make([]byte, dataLen)
	for i := 0; i < r.dataShards; i++ {
		copy(result[i*shardSize:], shards[i])
	}

	// Extract original data using size prefix
	if len(result) < 4 {
		return nil, fmt.Errorf("decoded data too short: %d bytes", len(result))
	}

	originalSize := binary.BigEndian.Uint32(result[:4])
	if originalSize > uint32(len(result)-4) {
		return nil, fmt.Errorf("invalid size prefix: %d bytes", originalSize)
	}

	return result[4 : 4+originalSize], nil
}
