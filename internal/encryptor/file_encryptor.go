package encryptor

import (
	"fmt"
	"runtime"

	"github.com/hambosto/go-encryption/internal/config"
	"github.com/schollz/progressbar/v3"
)

type FileEncryptor struct {
	chunkProcessor *ChunkProcessor
	bar            *progressbar.ProgressBar
	workers        int
}

func NewFileEncryptor(key []byte) (*FileEncryptor, error) {
	if len(key) != config.KeySize {
		return nil, fmt.Errorf("invalid key size: must be %d bytes", config.KeySize)
	}

	chunkProcessor, err := NewChunkProcessor(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create chunk processor: %w", err)
	}

	return &FileEncryptor{
		chunkProcessor: chunkProcessor,
		workers:        runtime.NumCPU(),
	}, nil
}

func (f *FileEncryptor) GetNonce() ([]byte, []byte) {
	return f.chunkProcessor.serpentCipher.GetNonce(), f.chunkProcessor.chaCha20Cipher.GetNonce()
}
