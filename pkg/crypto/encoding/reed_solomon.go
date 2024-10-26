package encoding

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/hambosto/go-encryption/pkg/crypto/config"
	"github.com/vivint/infectious"
)

const (
	MinShards   = 1
	MaxShards   = 256
	MinDataSize = 1
	MaxDataSize = 1 << 32
)

type ReedSolomon struct {
	fec         *infectious.FEC
	dataShards  int
	totalShards int
}

func NewReedSolomon(dataShards, parityShards int) (*ReedSolomon, error) {
	if dataShards < MinShards || dataShards > MaxShards {
		return nil, fmt.Errorf("invalid number of data shards: must be between %d and %d", MinShards, config.DataShards)
	}

	if parityShards < MinShards || parityShards > MaxShards {
		return nil, fmt.Errorf("invalid number of parity shards: must be between %d and %d", MinShards, MaxShards)
	}

	totalShards := dataShards + parityShards
	if totalShards > MaxShards {
		return nil, fmt.Errorf("total number shards (%d) exceeds maximum allowed (%d)", totalShards, MaxShards)
	}

	fec, err := infectious.NewFEC(dataShards, totalShards)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize FEC: %w", err)
	}

	return &ReedSolomon{
		fec:         fec,
		dataShards:  dataShards,
		totalShards: totalShards,
	}, nil
}

func (r *ReedSolomon) Encode(data []byte) ([]byte, error) {
	if len(data) < MinDataSize || len(data) > MaxDataSize {
		return nil, fmt.Errorf("invalid data size: must be between %d and %d", MinDataSize, MaxDataSize)
	}

	sizePrefix := make([]byte, 4)
	binary.BigEndian.PutUint32(sizePrefix, uint32(len(data)))
	paddedData := append(sizePrefix, data...)

	buffer := &bytes.Buffer{}
	if err := r.fec.Encode(paddedData, func(s infectious.Share) { buffer.Write(s.Data) }); err != nil {
		return nil, fmt.Errorf("failed to encode data: %w", err)
	}

	return buffer.Bytes(), nil
}

func (r *ReedSolomon) Decode(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("invalid data size: must be between %d and %d", MinDataSize, MaxDataSize)
	}

	if len(data)%r.totalShards != 0 {
		return nil, fmt.Errorf("encoded data length (%d) is not divisible by total shards (%d)", len(data), r.totalShards)
	}

	shareSize := len(data) / r.totalShards
	shares := make([]infectious.Share, r.totalShards)
	for i := 0; i < r.totalShards; i++ {
		start := i * shareSize
		shares[i] = infectious.Share{
			Data:   data[start : start+shareSize],
			Number: i,
		}
	}

	decoded, err := r.fec.Decode(nil, shares[:r.dataShards])
	if err != nil {
		return nil, fmt.Errorf("failed to decode data: %w", err)
	}

	if len(decoded) < 4 {
		return nil, fmt.Errorf("decoded data too short: %d bytes", len(decoded))
	}

	originalSize := binary.BigEndian.Uint32(decoded[:4])
	if originalSize > uint32(len(decoded)-4) {
		return nil, fmt.Errorf("invalid size prefix: %d bytes", len(decoded))
	}

	return decoded[4 : 4+originalSize], nil
}
