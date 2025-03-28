package encoding

import (
	"fmt"
)

func validateConfig(config ReedSolomonConfig) error {
	if config.DataShards <= 0 {
		return fmt.Errorf("data shards must be positive")
	}
	if config.ParityShards <= 0 {
		return fmt.Errorf("parity shards must be positive")
	}
	return nil
}

func validateData(data []byte) error {
	if len(data) == 0 || len(data) > maxDataSize {
		return fmt.Errorf("invalid data size: must be between 1 and %d bytes", maxDataSize)
	}
	return nil
}

func validateEncodedData(data []byte, totalShards int) error {
	if len(data) == 0 || len(data)%totalShards != 0 {
		return fmt.Errorf("invalid encoded data size")
	}
	return nil
}
