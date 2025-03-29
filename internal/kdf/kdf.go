package kdf

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/argon2"
)

type KDF interface {
	DeriveKey(password, salt []byte) ([]byte, error)
	GenerateSalt() ([]byte, error)
	GetConfig() *Config
}

type kdf struct {
	config *Config
}

func newKDF(config *Config) KDF {
	return &kdf{config: config}
}

func (k *kdf) DeriveKey(password, salt []byte) ([]byte, error) {
	if err := k.validateInput(password, salt); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	key := argon2.IDKey(
		password,
		salt,
		k.config.GetTimeCost(),
		k.config.GetMemory(),
		k.config.GetThreads(),
		k.config.GetKeyLength(),
	)

	return key, nil
}

func (k *kdf) GenerateSalt() ([]byte, error) {
	salt := make([]byte, k.config.GetSaltLength())
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}

func (k *kdf) GetConfig() *Config {
	return k.config.Clone()
}

func (k *kdf) validateInput(password, salt []byte) error {
	if len(password) == 0 {
		return ErrEmptyPassword
	}

	if len(salt) != int(k.config.GetSaltLength()) {
		return fmt.Errorf("%w: expected %d, got %d",
			ErrInvalidSaltLength,
			k.config.GetSaltLength(),
			len(salt))
	}

	return nil
}
