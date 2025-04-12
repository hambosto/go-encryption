package kdf

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/argon2"
)

// Deriver is the main interface for key derivation operations
type Deriver interface {
	DeriveKey(password, salt []byte) ([]byte, error)
	GenerateSalt() ([]byte, error)
	GetParameters() Parameters
}

// argon2Deriver implements the Deriver interface using Argon2id
type argon2Deriver struct {
	params Parameters
}

// DeriveKey generates a key from a password and salt
func (d *argon2Deriver) DeriveKey(password, salt []byte) ([]byte, error) {
	if len(password) == 0 {
		return nil, ErrEmptyPassword
	}

	if uint32(len(salt)) != d.params.SaltBytes {
		return nil, fmt.Errorf("%w: expected %d, got %d",
			ErrInvalidSaltLength,
			d.params.SaltBytes,
			len(salt))
	}

	key := argon2.IDKey(
		password,
		salt,
		d.params.Iterations,
		d.params.MemoryMB*1024, // Convert MB to KB
		d.params.Parallelism,
		d.params.KeyBytes,
	)

	return key, nil
}

// GenerateSalt creates a cryptographically secure random salt
func (d *argon2Deriver) GenerateSalt() ([]byte, error) {
	salt := make([]byte, d.params.SaltBytes)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random salt: %w", err)
	}
	return salt, nil
}

// GetParameters returns a copy of the current parameters
func (d *argon2Deriver) GetParameters() Parameters {
	return d.params
}
