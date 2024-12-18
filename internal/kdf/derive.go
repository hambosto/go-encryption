package kdf

import (
	"github.com/hambosto/go-encryption/internal/config"
	"golang.org/x/crypto/argon2"
)

func Derive(password []byte, salt []byte) ([]byte, error) {
	if err := validatePassword(password); err != nil {
		return nil, err
	}

	if err := validateSalt(salt); err != nil {
		return nil, err
	}

	key := argon2.Key(
		password,
		salt,
		argonTimeCost,
		argonMemory,
		argonThreads,
		config.KeySize,
	)

	return key, nil
}
