package kdf

import (
	"fmt"

	"github.com/hambosto/go-encryption/internal/constants"
)

func validatePassword(password []byte) error {
	if len(password) == 0 {
		return fmt.Errorf("password cannot be empty")
	}
	return nil
}

func validateSalt(salt []byte) error {
	if len(salt) != constants.SaltSize {
		return fmt.Errorf("salt must be %d bytes", constants.SaltSize)
	}
	return nil
}
