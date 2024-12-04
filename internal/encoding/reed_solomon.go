package encoding

import (
	"fmt"

	"github.com/hambosto/go-encryption/internal/config"
	"github.com/vivint/infectious"
)

type ReedSolomonEncoder struct {
	fec         *infectious.FEC
	dataShards  int
	totalShards int
}

func NewReedSolomonEncoder(dataShards, parityShards int) (*ReedSolomonEncoder, error) {
	if err := validateConfig(ReedSolomonConfig{DataShards: dataShards, ParityShards: parityShards}); err != nil {
		return nil, err
	}

	totalShards := dataShards + parityShards
	fec, err := infectious.NewFEC(config.DataShards, totalShards)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize FEC: %w", err)
	}

	return &ReedSolomonEncoder{
		fec:         fec,
		dataShards:  config.DataShards,
		totalShards: totalShards,
	}, nil
}

func (r *ReedSolomonEncoder) Encode(data []byte) ([]byte, error) {
	paddedData, err := prepareDataForEncoding(data)
	if err != nil {
		return nil, err
	}

	return encodeWithFEC(r.fec, paddedData)
}

func (r *ReedSolomonEncoder) Decode(data []byte) ([]byte, error) {
	if err := validateEncodedData(data, r.totalShards); err != nil {
		return nil, err
	}

	shares, err := reconstructShares(data, r.totalShards)
	if err != nil {
		return nil, err
	}

	decoded, err := r.fec.Decode(nil, shares[:r.dataShards])
	if err != nil {
		return nil, fmt.Errorf("failed to decode data: %w", err)
	}

	return extractOriginalData(decoded)
}
