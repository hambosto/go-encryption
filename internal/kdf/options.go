package kdf

type Option func(*Config) error

func WithMemory(memory uint32) Option {
	return func(c *Config) error {
		c.Memory = memory
		return nil
	}
}

func WithTimeCost(timeCost uint32) Option {
	return func(c *Config) error {
		c.TimeCost = timeCost
		return nil
	}
}

func WithThreads(threads uint8) Option {
	return func(c *Config) error {
		c.Threads = threads
		return nil
	}
}

func WithKeyLength(keyLength uint32) Option {
	return func(c *Config) error {
		c.KeyLength = keyLength
		return nil
	}
}

func WithSaltLength(saltLength uint32) Option {
	return func(c *Config) error {
		c.SaltLength = saltLength
		return nil
	}
}
