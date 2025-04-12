package kdf

import (
	"errors"
	"fmt"
)

// Common errors
var (
	ErrEmptyPassword     = errors.New("password cannot be empty")
	ErrInvalidSaltLength = errors.New("salt length doesn't match configuration")
	ErrInvalidParameters = errors.New("invalid parameters")
)

// Parameters holds the configuration for the key derivation function
type Parameters struct {
	MemoryMB    uint32 // Memory usage in MB
	Iterations  uint32 // Time cost
	Parallelism uint8  // Number of threads
	KeyBytes    uint32 // Output key length in bytes
	SaltBytes   uint32 // Salt length in bytes
}

// DefaultParameters returns the recommended secure parameters
func DefaultParameters() Parameters {
	return Parameters{
		MemoryMB:    64, // 64MB
		Iterations:  4,  // 4 iterations
		Parallelism: 4,  // 4 threads
		KeyBytes:    64, // 64 byte key
		SaltBytes:   32, // 32 byte salt
	}
}

// MinimumParameters returns the minimum acceptable parameters
func MinimumParameters() Parameters {
	return Parameters{
		MemoryMB:    8,  // 8MB minimum
		Iterations:  1,  // At least 1 iteration
		Parallelism: 1,  // At least 1 thread
		KeyBytes:    16, // At least 16 byte key
		SaltBytes:   16, // At least 16 byte salt
	}
}

// Validate checks if parameters are acceptable
func (p Parameters) Validate() error {
	min := MinimumParameters()

	if p.MemoryMB < min.MemoryMB {
		return fmt.Errorf("%w: memory must be at least %d MB", ErrInvalidParameters, min.MemoryMB)
	}
	if p.Iterations < min.Iterations {
		return fmt.Errorf("%w: iterations must be at least %d", ErrInvalidParameters, min.Iterations)
	}
	if p.Parallelism < min.Parallelism {
		return fmt.Errorf("%w: parallelism must be at least %d", ErrInvalidParameters, min.Parallelism)
	}
	if p.KeyBytes < min.KeyBytes {
		return fmt.Errorf("%w: key length must be at least %d bytes", ErrInvalidParameters, min.KeyBytes)
	}
	if p.SaltBytes < min.SaltBytes {
		return fmt.Errorf("%w: salt length must be at least %d bytes", ErrInvalidParameters, min.SaltBytes)
	}

	return nil
}
