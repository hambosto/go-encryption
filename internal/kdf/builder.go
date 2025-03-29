package kdf

type Builder struct {
	config *Config
}

func NewBuilder() *Builder {
	return &Builder{
		config: NewConfig(),
	}
}

func (b *Builder) WithMemory(memory uint32) *Builder {
	b.config.memory = memory
	return b
}

func (b *Builder) WithTimeCost(timeCost uint32) *Builder {
	b.config.timeCost = timeCost
	return b
}

func (b *Builder) WithThreads(threads uint8) *Builder {
	b.config.threads = threads
	return b
}

func (b *Builder) WithKeyLength(keyLength uint32) *Builder {
	b.config.keyLength = keyLength
	return b
}

func (b *Builder) WithSaltLength(saltLength uint32) *Builder {
	b.config.saltLength = saltLength
	return b
}

func (b *Builder) Build() (KDF, error) {
	if err := b.config.Validate(); err != nil {
		return nil, err
	}

	return newKDF(b.config.Clone()), nil
}
