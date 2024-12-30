package encoding

import (
	"fmt"

	"github.com/klauspost/reedsolomon"
)

type ReedSolomon struct {
	encoder      reedsolomon.Encoder
	dataShards   int
	parityShards int
}

func NewReedSolomon(dataShards, parityShards int) (*ReedSolomon, error) {
	if err := validateShards(dataShards, parityShards); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	enc, err := reedsolomon.New(dataShards, parityShards)
	if err != nil {
		return nil, fmt.Errorf("failed to create encoder: %w", err)
	}

	return &ReedSolomon{
		encoder:      enc,
		dataShards:   dataShards,
		parityShards: parityShards,
	}, nil
}
