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
	memory     uint32
	timeCost   uint32
	threads    uint8
	keyLength  uint32
	saltLength uint32
}

func NewConfig() *Config {
	return &Config{
		memory:     64 * 1024,
		timeCost:   4,
		threads:    4,
		keyLength:  64,
		saltLength: 32,
	}
}

func (c *Config) Clone() *Config {
	return &Config{
		memory:     c.memory,
		timeCost:   c.timeCost,
		threads:    c.threads,
		keyLength:  c.keyLength,
		saltLength: c.saltLength,
	}
}

func (c *Config) Validate() error {
	if c.memory < 8*1024 {
		return fmt.Errorf("%w: memory must be at least 8MB", ErrInvalidConfig)
	}
	if c.timeCost < 1 {
		return fmt.Errorf("%w: time cost must be at least 1", ErrInvalidConfig)
	}
	if c.threads < 1 {
		return fmt.Errorf("%w: threads must be at least 1", ErrInvalidConfig)
	}
	if c.keyLength < 16 {
		return fmt.Errorf("%w: key length must be at least 16 bytes", ErrInvalidConfig)
	}
	if c.saltLength < 16 {
		return fmt.Errorf("%w: salt length must be at least 16 bytes", ErrInvalidConfig)
	}
	return nil
}

func (c *Config) GetMemory() uint32 {
	return c.memory
}

func (c *Config) GetTimeCost() uint32 {
	return c.timeCost
}

func (c *Config) GetThreads() uint8 {
	return c.threads
}

func (c *Config) GetKeyLength() uint32 {
	return c.keyLength
}

func (c *Config) GetSaltLength() uint32 {
	return c.saltLength
}
