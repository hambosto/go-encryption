package encryptor

import (
	"fmt"
	"runtime"

	"github.com/schollz/progressbar/v3"
)

type FileEncryptor struct {
	chunkProcessor *ChunkProcessor
	bar            *progressbar.ProgressBar
	workers        int
}

func NewFileEncryptor(key []byte) (*FileEncryptor, error) {
	if len(key) != 64 {
		return nil, fmt.Errorf("invalid key size: must be %d bytes", 64)
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
	return f.chunkProcessor.aesCipher.GetNonce(), f.chunkProcessor.chaCha20Cipher.GetNonce()
}
