package kdf

import (
	"fmt"
)

func validatePassword(password []byte) error {
	if len(password) == 0 {
		return fmt.Errorf("password cannot be empty")
	}
	return nil
}

func validateSalt(salt []byte) error {
	if len(salt) != 32 {
		return fmt.Errorf("salt must be %d bytes", 32)
	}
	return nil
}
