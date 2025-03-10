package worker

import (
	"fmt"
	"runtime"

	"github.com/hambosto/go-encryption/internal/processor"
	"github.com/schollz/progressbar/v3"
)

type FileProcessor struct {
	chunkProcessor *processor.Processor
	bar            *progressbar.ProgressBar
	workers        int
}

func NewFileProcessor(key []byte, isEncryption bool) (*FileProcessor, error) {
	if len(key) != 64 {
		return nil, fmt.Errorf("invalid key size: must be %d bytes", 64)
	}

	chunkProcessor, err := processor.NewProcessor(key, isEncryption)
	if err != nil {
		return nil, err
	}

	return &FileProcessor{
		chunkProcessor: chunkProcessor,
		workers:        runtime.NumCPU(),
	}, nil
}

func (f *FileProcessor) GetAesNonce() []byte {
	return f.chunkProcessor.AESCipher.GetNonce()
}

func (f *FileProcessor) GetChaCha20Nonce() []byte {
	return f.chunkProcessor.ChaCha20Cipher.GetNonce()
}

func (f *FileProcessor) SetAesNonce(aesNonce []byte) error {
	return f.chunkProcessor.AESCipher.SetNonce(aesNonce)
}

func (f *FileProcessor) SetChaCha20Nonce(aesNonce []byte) error {
	return f.chunkProcessor.ChaCha20Cipher.SetNonce(aesNonce)
}
