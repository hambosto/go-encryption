package kdf

import (
	"crypto/rand"
	"fmt"

	"github.com/hambosto/go-encryption/internal/constants"
	"golang.org/x/crypto/argon2"
)

func Derive(password []byte, salt []byte) ([]byte, error) {
	if len(password) == 0 {
		return nil, fmt.Errorf("password cannot be empty")
	}

	if len(salt) != constants.SaltSize {
		return nil, fmt.Errorf("salt must be %d bytes", constants.SaltSize)
	}

	key := argon2.Key(
		password,
		salt,
		4,
		64*1024,
		4,
		constants.KeySize,
	)

	return key, nil
}

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, constants.SaltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}
