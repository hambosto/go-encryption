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
	var err []error

	if c.Memory < 8*1024 {
		err = append(err, fmt.Errorf("memory must be at least 8MB"))
	}
	if c.TimeCost < 1 {
		err = append(err, fmt.Errorf("time cost must be at least 1"))
	}
	if c.Threads < 1 {
		err = append(err, fmt.Errorf("threads must be at least 1"))
	}
	if c.KeyLength < 16 {
		err = append(err, fmt.Errorf("key length must be at least 16 bytes"))
	}
	if c.SaltLength < 16 {
		err = append(err, fmt.Errorf("salt length must be at least 16 bytes"))
	}

	if len(err) > 0 {
		return fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}
	return nil
}
