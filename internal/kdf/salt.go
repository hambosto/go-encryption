package kdf

import (
	"crypto/rand"
	"fmt"

	"github.com/hambosto/go-encryption/internal/config"
)

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, config.SaltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}
