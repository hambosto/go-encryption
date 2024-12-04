package decryptor

const (
	MaxChunkSize          = 1024 * 1024
	MaxEncryptedChunkSize = ((MaxChunkSize + (MaxChunkSize / 10) + 16 + 4 + (4 - 1)) / 4) * (4 + 10)
)
