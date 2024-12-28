package decryptor

import (
	"fmt"
	"runtime"

	"github.com/hambosto/go-encryption/internal/config"
	"github.com/schollz/progressbar/v3"
)

type FileDecryptor struct {
	chunkProcessor *ChunkProcessor
	bar            *progressbar.ProgressBar
	workers        int
}

func NewFileDecryptor(key []byte) (*FileDecryptor, error) {
	if len(key) < config.KeySize {
		return nil, fmt.Errorf("invalid key size: must be %d bytes", config.KeySize)
	}

	chunkProcessor, err := NewChunkProcessor(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create chunk processor: %w", err)
	}

	return &FileDecryptor{
		chunkProcessor: chunkProcessor,
		workers:        runtime.NumCPU(),
	}, nil
}

func (f *FileDecryptor) SetNonce(aesNonce, chacha20Nonce []byte) error {
	if err := f.chunkProcessor.aesCipher.SetNonce(aesNonce); err != nil {
		return fmt.Errorf("failed to set aes nonce: %w", err)
	}

	if err := f.chunkProcessor.chaCha20Cipher.SetNonce(chacha20Nonce); err != nil {
		return fmt.Errorf("failed to set chacha20 nonce: %w", err)
	}

	return nil
}
