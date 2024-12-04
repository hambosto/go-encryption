package kdf

import (
	"fmt"

	"github.com/hambosto/go-encryption/internal/config"
)

func validatePassword(password []byte) error {
	if len(password) == 0 {
		return fmt.Errorf("password cannot be empty")
	}
	return nil
}

func validateSalt(salt []byte) error {
	if len(salt) != config.SaltSize {
		return fmt.Errorf("salt must be %d bytes", config.SaltSize)
	}
	return nil
}
