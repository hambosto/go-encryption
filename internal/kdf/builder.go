package kdf

// Builder is used to configure and create a Deriver
type Builder struct {
	params Parameters
}

// NewBuilder creates a Builder with default parameters
func NewBuilder() *Builder {
	return &Builder{
		params: DefaultParameters(),
	}
}

// WithMemory sets the memory usage in MB
func (b *Builder) WithMemory(memoryMB uint32) *Builder {
	b.params.MemoryMB = memoryMB
	return b
}

// WithIterations sets the time cost parameter
func (b *Builder) WithIterations(iterations uint32) *Builder {
	b.params.Iterations = iterations
	return b
}

// WithParallelism sets the number of threads to use
func (b *Builder) WithParallelism(threads uint8) *Builder {
	b.params.Parallelism = threads
	return b
}

// WithKeyLength sets the output key length in bytes
func (b *Builder) WithKeyLength(keyBytes uint32) *Builder {
	b.params.KeyBytes = keyBytes
	return b
}

// WithSaltLength sets the salt length in bytes
func (b *Builder) WithSaltLength(saltBytes uint32) *Builder {
	b.params.SaltBytes = saltBytes
	return b
}

// WithParameters sets all parameters at once
func (b *Builder) WithParameters(params Parameters) *Builder {
	b.params = params
	return b
}

// Build creates and returns a new Deriver with the configured parameters
func (b *Builder) Build() (Deriver, error) {
	if err := b.params.Validate(); err != nil {
		return nil, err
	}

	return &argon2Deriver{
		params: b.params,
	}, nil
}
