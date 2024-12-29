package operations

import (
	"fmt"
	"runtime"

	"github.com/schollz/progressbar/v3"
)

type FileProcessor struct {
	chunkProcessor *ChunkProcessor
	bar            *progressbar.ProgressBar
	workers        int
}

func NewFileProcessor(key []byte, isEncryption bool) (*FileProcessor, error) {
	if len(key) != 64 {
		return nil, fmt.Errorf("invalid key size: must be %d bytes", 64)
	}

	chunkProcessor, err := NewChunkProcessor(key, isEncryption)
	if err != nil {
		return nil, err
	}

	return &FileProcessor{
		chunkProcessor: chunkProcessor,
		workers:        runtime.NumCPU(),
	}, nil
}

func (f *FileProcessor) GetNonce() ([]byte, []byte) {
	return f.chunkProcessor.aesCipher.GetNonce(), f.chunkProcessor.chaCha20Cipher.GetNonce()
}

func (f *FileProcessor) SetNonce(aesNonce, chacha20Nonce []byte) error {
	if err := f.chunkProcessor.aesCipher.SetNonce(aesNonce); err != nil {
		return fmt.Errorf("failed to set aes nonce: %w", err)
	}

	if err := f.chunkProcessor.chaCha20Cipher.SetNonce(chacha20Nonce); err != nil {
		return fmt.Errorf("failed to set chacha20 nonce: %w", err)
	}

	return nil
}
