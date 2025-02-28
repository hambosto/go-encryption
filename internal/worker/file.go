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

func (f *FileProcessor) GetPrimaryCipherNonce() []byte {
	return f.chunkProcessor.GetPrimaryCipher().GetNonce()
}

func (f *FileProcessor) GetSecondaryCipherNonce() []byte {
	return f.chunkProcessor.GetSecondaryCipher().GetNonce()
}

func (f *FileProcessor) SetPrimaryCipherNonce(nonce []byte) error {
	return f.chunkProcessor.GetPrimaryCipher().SetNonce(nonce)
}

func (f *FileProcessor) SetSecondaryCipherNonce(nonce []byte) error {
	return f.chunkProcessor.GetSecondaryCipher().SetNonce(nonce)
}
