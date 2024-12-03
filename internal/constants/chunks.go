package constants

// Chunk and size-related constants
const (
	ChunkSizeHeader       = 4
	MaxChunkSize          = 1024 * 1024
	MaxCompressedSize     = MaxChunkSize + (MaxChunkSize / 10)
	MaxEncryptedSize      = MaxCompressedSize + TagSize + ChunkSizeHeader
	MaxEncryptedChunkSize = ((MaxEncryptedSize + (DataShards - 1)) / DataShards) * (DataShards + ParityShards)
)
