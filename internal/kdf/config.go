package kdf

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyPassword     = errors.New("password cannot be empty")
	ErrInvalidSaltLength = errors.New("invalid salt length")
	ErrInvalidConfig     = errors.New("invalid configuration")
)

type Config struct {
	Memory     uint32
	TimeCost   uint32
	Threads    uint8
	KeyLength  uint32
	SaltLength uint32
}

func (c *Config) Clone() *Config {
	return &Config{
		Memory:     c.Memory,
		TimeCost:   c.TimeCost,
		Threads:    c.Threads,
		KeyLength:  c.KeyLength,
		SaltLength: c.SaltLength,
	}
}

func (c *Config) validate() error {
	if c.Memory < 8*1024 {
		return fmt.Errorf("%w: memory must be at least 8MB", ErrInvalidConfig)
	}
	if c.TimeCost < 1 {
		return fmt.Errorf("%w: time cost must be at least 1", ErrInvalidConfig)
	}
	if c.Threads < 1 {
		return fmt.Errorf("%w: threads must be at least 1", ErrInvalidConfig)
	}
	if c.KeyLength < 16 {
		return fmt.Errorf("%w: key length must be at least 16 bytes", ErrInvalidConfig)
	}
	if c.SaltLength < 16 {
		return fmt.Errorf("%w: salt length must be at least 16 bytes", ErrInvalidConfig)
	}
	return nil
}
