package kdf

import "fmt"

type Config struct {
	// Memory in KiB
	Memory uint32
	// Number of iterations
	TimeCost uint32
	// Degree of parallelism
	Threads uint8
	// Key length in bytes
	KeyLength uint32
	// Salt length in bytes
	SaltLength uint32
}

func DefaultConfig() *Config {
	return &Config{
		Memory:     64 * 1024, // 64 MB
		TimeCost:   4,
		Threads:    4,
		KeyLength:  64,
		SaltLength: 32,
	}
}

type ConfigBuilder struct {
	config *Config
}

func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: DefaultConfig(),
	}
}

func (b *ConfigBuilder) WithMemory(memory uint32) *ConfigBuilder {
	b.config.Memory = memory
	return b
}

func (b *ConfigBuilder) WithTimeCost(timeCost uint32) *ConfigBuilder {
	b.config.TimeCost = timeCost
	return b
}

func (b *ConfigBuilder) WithThreads(threads uint8) *ConfigBuilder {
	b.config.Threads = threads
	return b
}

func (b *ConfigBuilder) WithKeyLength(keyLength uint32) *ConfigBuilder {
	b.config.KeyLength = keyLength
	return b
}

func (b *ConfigBuilder) WithSaltLength(saltLength uint32) *ConfigBuilder {
	b.config.SaltLength = saltLength
	return b
}

func (b *ConfigBuilder) Build() (*Config, error) {
	if err := b.config.validate(); err != nil {
		return nil, err
	}
	return b.config, nil
}

func (c *Config) validate() error {
	if c.Memory < 8*1024 {
		return fmt.Errorf("memory must be at least 8MB")
	}
	if c.TimeCost < 1 {
		return fmt.Errorf("time cost must be at least 1")
	}
	if c.Threads < 1 {
		return fmt.Errorf("threads must be at least 1")
	}
	if c.KeyLength < 16 {
		return fmt.Errorf("key length must be at least 16 bytes")
	}
	if c.SaltLength < 16 {
		return fmt.Errorf("salt length must be at least 16 bytes")
	}
	return nil
}
