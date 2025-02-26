package kdf

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/argon2"
)

type KDF struct {
	config *Config
}

func New(config *Config) (*KDF, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}
	return &KDF{config: config}, nil
}

func NewWithDefaults() *KDF {
	return &KDF{config: DefaultConfig()}
}

func (k *KDF) GenerateSalt() ([]byte, error) {
	salt := make([]byte, k.config.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}

func (k *KDF) DeriveKey(password, salt []byte) ([]byte, error) {
	if err := k.validateInput(password, salt); err != nil {
		return nil, err
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

func (k *KDF) validateInput(password, salt []byte) error {
	if len(password) == 0 {
		return fmt.Errorf("password cannot be empty")
	}

	if len(salt) != int(k.config.SaltLength) {
		return fmt.Errorf("salt length must be %d bytes", k.config.SaltLength)
	}

	return nil
}
