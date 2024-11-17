package config

// Constants related to encryption configurations
const (
	// KeySize is the size of the encryption key in bytes.
	KeySize = 64
	// SaltSize is the size of the salt used for encryption in bytes.
	SaltSize = 32
	// NonceSize is the size of the nonce (number used once) in bytes.
	NonceSize = 12
	// TagSize is the size of the authentication tag in bytes.
	TagSize = 16
	// NonceSizeX is the size of the nonce (number used once) in bytes.
	NonceSizeX = 24
)

// Constants for data encoding and redundancy (for Reed-Solomon or similar algorithms)
const (
	// DataShards is the number of data shards used in erasure coding.
	DataShards = 4
	// ParityShards is the number of parity shards used for error correction.
	ParityShards = 10
)

// Constants related to chunked data sizes and processing
const (
	// ChunkSizeHeader is the size of the chunk header in bytes.
	ChunkSizeHeader = 4
	// MaxChunkSize is the maximum size of a chunk in bytes (1 MB).
	MaxChunkSize = 1024 * 1024
)

// Derived constants based on chunk size, compression, and encryption overhead
const (
	// MaxCompressedSize is the maximum size of compressed data for a chunk,
	// allowing for up to 10% compression overhead.
	MaxCompressedSize = MaxChunkSize + (MaxChunkSize / 10)
	// MaxEncryptedSize is the maximum size of encrypted data, accounting for
	// tag size and chunk size header.
	MaxEncryptedSize = MaxCompressedSize + TagSize + ChunkSizeHeader
	// MaxEncryptedChunkSize is the maximum size of an encrypted chunk,
	// adjusted for data and parity shards used in erasure coding.
	MaxEncryptedChunkSize = ((MaxEncryptedSize + (DataShards - 1)) / DataShards) * (DataShards + ParityShards)
)
