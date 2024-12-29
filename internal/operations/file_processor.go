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

func (f *FileProcessor) GetAesNonce() []byte {
	return f.chunkProcessor.aesCipher.GetNonce()
}

func (f *FileProcessor) GetChaCha20Nonce() []byte {
	return f.chunkProcessor.chaCha20Cipher.GetNonce()
}

func (f *FileProcessor) SetAesNonce(aesNonce []byte) error {
	return f.chunkProcessor.aesCipher.SetNonce(aesNonce)
}

func (f *FileProcessor) SetChaCha20Nonce(aesNonce []byte) error {
	return f.chunkProcessor.chaCha20Cipher.SetNonce(aesNonce)
}
