package kdf

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/argon2"
)

type KDF interface {
	DeriveKey(password, salt []byte) ([]byte, error)
	GenerateSalt() ([]byte, error)
	Config() *Config
}

type kdf struct {
	config *Config
}

func New(opts ...Option) (KDF, error) {
	config := &Config{
		Memory:     64 * 1024, // 64 MB
		TimeCost:   4,
		Threads:    4,
		KeyLength:  64,
		SaltLength: 32,
	}

	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &kdf{config: config}, nil
}

func (k *kdf) DeriveKey(password, salt []byte) ([]byte, error) {
	if err := k.validateInput(password, salt); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	key := argon2.IDKey(
		password,
		salt,
		k.config.TimeCost,
		k.config.Memory,
		k.config.Threads,
		k.config.KeyLength,
	)

	return key, nil
}

func (k *kdf) GenerateSalt() ([]byte, error) {
	salt := make([]byte, k.config.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}

func (k *kdf) Config() *Config {
	return k.config.Clone()
}

func (k *kdf) validateInput(password, salt []byte) error {
	if len(password) == 0 {
		return ErrEmptyPassword
	}

	if len(salt) != int(k.config.SaltLength) {
		return fmt.Errorf("%w: expected %d, got %d", ErrInvalidSaltLength, k.config.SaltLength, len(salt))
	}

	return nil
}
