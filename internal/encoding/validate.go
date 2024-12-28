package encoding

import "fmt"

func validateShards(data, parity int) error {
	if data <= 0 || parity <= 0 {
		return fmt.Errorf("shard counts must be positive")
	}

	if data > maxShards || parity > maxShards {
		return fmt.Errorf("shard count exceeds maximum of %d", maxShards)
	}

	if data+parity > maxShards {
		return fmt.Errorf("total shard count exceeds maximum of %d", maxShards)
	}

	return nil
}
