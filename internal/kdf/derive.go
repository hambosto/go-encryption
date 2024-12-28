package kdf

import (
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
		64,
	)

	return key, nil
}
