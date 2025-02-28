package encoding

import (
	"fmt"

	"github.com/klauspost/reedsolomon"
)

type ReedSolomonEncoder struct {
	dataShards   int
	parityShards int
	encoder      reedsolomon.Encoder
}

func NewReedSolomonEncoder(config Config) (*ReedSolomonEncoder, error) {
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	enc, err := reedsolomon.New(config.DataShards, config.ParityShards)
	if err != nil {
		return nil, fmt.Errorf("failed to create reed-solomon encoder: %w", err)
	}

	return &ReedSolomonEncoder{
		dataShards:   config.DataShards,
		parityShards: config.ParityShards,
		encoder:      enc,
	}, nil
}

func validateConfig(config Config) error {
	if config.DataShards <= 0 {
		return NewValidationError("data shards must be positive")
	}
	if config.ParityShards <= 0 {
		return NewValidationError("parity shards must be positive")
	}
	return nil
}
