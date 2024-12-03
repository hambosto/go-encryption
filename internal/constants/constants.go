package constants

const (
	KeySize    = 64
	SaltSize   = 32
	NonceSize  = 12
	TagSize    = 16
	NonceSizeX = 24
)

const (
	DataShards   = 4
	ParityShards = 10
)

const (
	ChunkSizeHeader = 4
	MaxChunkSize    = 1024 * 1024
)

const (
	MaxCompressedSize     = MaxChunkSize + (MaxChunkSize / 10)
	MaxEncryptedSize      = MaxCompressedSize + TagSize + ChunkSizeHeader
	MaxEncryptedChunkSize = ((MaxEncryptedSize + (DataShards - 1)) / DataShards) * (DataShards + ParityShards)
)
