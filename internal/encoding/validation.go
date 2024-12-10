package encoding

import "fmt"

func validateConfig(dataShards, parityShards int) error {
	if dataShards < MinShards || dataShards > MaxShards {
		return fmt.Errorf("invalid number of data shards: must be between %d and %d", MinShards, MaxShards)
	}

	if parityShards < MinShards || parityShards > MaxShards {
		return fmt.Errorf("invalid number of parity shards: must be between %d and %d", MinShards, MaxShards)
	}

	totalShards := dataShards + parityShards
	if totalShards > MaxShards {
		return fmt.Errorf("total number shards (%d) exceeds maximum allowed (%d)", totalShards, MaxShards)
	}

	return nil
}

func validateEncodedData(data []byte, totalShards int) error {
	if len(data) == 0 {
		return fmt.Errorf("invalid data size: must be between %d and %d", MinDataSize, MaxDataSize)
	}

	if len(data)%totalShards != 0 {
		return fmt.Errorf("encoded data length (%d) is not divisible by total shards (%d)", len(data), totalShards)
	}

	return nil
}

func validateDataSize(data []byte) error {
	if len(data) < MinDataSize || len(data) > MaxDataSize {
		return fmt.Errorf("invalid data size: must be between %d and %d", MinDataSize, MaxDataSize)
	}

	return nil
}
