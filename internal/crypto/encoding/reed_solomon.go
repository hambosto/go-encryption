package encoding

import (
	"bytes"
	"fmt"

	"github.com/vivint/infectious"
)

// Constants for validation
const (
	MinShards   = 1       // Minimum number of shards (data or parity)
	MaxShards   = 256     // Maximum number of shards (data + parity)
	MinDataSize = 1       // Minimum data size for encoding
	MaxDataSize = 1 << 32 // Maximum data size (4GB limit)
)

// ReedSolomonCodec implements Reed-Solomon encoding and decoding operations.
type ReedSolomonCodec struct {
	fec         *infectious.FEC // Forward Error Correction instance
	dataShards  int             // Number of data shards
	totalShards int             // Total number of shards (data + parity)
}

// Config holds the configuration parameters for ReedSolomonCodec.
type Config struct {
	DataShards   int // Number of data shards
	ParityShards int // Number of parity shards
}

// Validate checks if the configuration parameters are valid.
func (c *Config) Validate() error {
	if c.DataShards < MinShards || c.DataShards > MaxShards {
		return fmt.Errorf("invalid number of data shards: must be between %d and %d", MinShards, MaxShards)
	}
	if c.ParityShards < MinShards || c.ParityShards > MaxShards {
		return fmt.Errorf("invalid number of parity shards: must be between %d and %d", MinShards, MaxShards)
	}
	if total := c.DataShards + c.ParityShards; total > MaxShards {
		return fmt.Errorf("total number of shards (%d) exceeds maximum allowed (%d)", total, MaxShards)
	}
	return nil
}

// NewReedSolomonCodec initializes a new ReedSolomonCodec with the specified configuration.
func NewReedSolomonCodec(config Config) (*ReedSolomonCodec, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	fec, err := infectious.NewFEC(config.DataShards, config.DataShards+config.ParityShards)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize FEC: %w", err)
	}

	return &ReedSolomonCodec{
		fec:         fec,
		dataShards:  config.DataShards,
		totalShards: config.DataShards + config.ParityShards,
	}, nil
}

// calculatePadding determines the required padding size to make data length divisible by the required shares.
func (codec *ReedSolomonCodec) calculatePadding(dataLen int) int {
	if padding := dataLen % codec.fec.Required(); padding != 0 {
		return codec.fec.Required() - padding
	}
	return 0
}

// padData ensures the input data length is divisible by the number of required shares by adding padding if necessary.
func (codec *ReedSolomonCodec) padData(data []byte) []byte {
	padding := codec.calculatePadding(len(data))
	if padding > 0 {
		paddedData := make([]byte, len(data)+padding)
		copy(paddedData, data)
		return paddedData
	}
	return data
}

// Encode encodes the input data using Reed-Solomon encoding.
func (codec *ReedSolomonCodec) Encode(data []byte) ([]byte, error) {
	if err := codec.validateDataSize(len(data)); err != nil {
		return nil, err
	}

	paddedData := codec.padData(data)
	encodedData := make([]byte, 0, len(paddedData)*codec.totalShards/codec.dataShards)
	buffer := bytes.NewBuffer(encodedData)

	// Perform encoding using the FEC instance
	err := codec.fec.Encode(paddedData, func(share infectious.Share) {
		buffer.Write(share.Data)
	})
	if err != nil {
		return nil, fmt.Errorf("encoding failed: %w", err)
	}

	return buffer.Bytes(), nil
}

// Decode reconstructs the original data from the encoded data using Reed-Solomon decoding.
func (codec *ReedSolomonCodec) Decode(encoded []byte) ([]byte, error) {
	if err := codec.validateEncodedData(encoded); err != nil {
		return nil, err
	}

	shareSize := len(encoded) / codec.totalShards
	shares := make([]infectious.Share, codec.dataShards)

	// Create shares for decoding from the encoded data
	for i := 0; i < codec.dataShards; i++ {
		start := i * shareSize
		end := start + shareSize
		shares[i] = infectious.Share{
			Data:   encoded[start:end],
			Number: i,
		}
	}

	// Perform decoding
	decoded, err := codec.fec.Decode(nil, shares)
	if err != nil {
		return nil, fmt.Errorf("decoding failed: %w", err)
	}

	// Remove padding
	return bytes.TrimRight(decoded, "\x00"), nil
}

// validateDataSize checks if the input data size is within acceptable limits.
func (codec *ReedSolomonCodec) validateDataSize(size int) error {
	if size < MinDataSize {
		return fmt.Errorf("data size too small: minimum is %d bytes", MinDataSize)
	}
	if size > MaxDataSize {
		return fmt.Errorf("data size too large: maximum is %d bytes", MaxDataSize)
	}
	return nil
}

// validateEncodedData checks if the encoded data is valid for decoding.
func (codec *ReedSolomonCodec) validateEncodedData(encoded []byte) error {
	if len(encoded) == 0 {
		return fmt.Errorf("encoded data cannot be empty")
	}

	if len(encoded)%codec.totalShards != 0 {
		return fmt.Errorf("encoded data length (%d) is not divisible by total shards (%d)", len(encoded), codec.totalShards)
	}

	return nil
}

