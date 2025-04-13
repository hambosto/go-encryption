package worker

import (
	"fmt"
	"io"
	"runtime"

	"github.com/hambosto/go-encryption/internal/processor"
	"github.com/schollz/progressbar/v3"
)

// ChunkSize defines the size of data chunks to process
const ChunkSize = 1024 * 1024 // 1MB

// FileProcessor handles the concurrent processing of file chunks
type FileProcessor struct {
	chunkProcessor *processor.ChunkProcessor
	progressBar    *progressbar.ProgressBar
	workerCount    int
}

// NewFileProcessor creates a new file processor with the given key and operation mode
func NewFileProcessor(key []byte, isEncryption bool) (*FileProcessor, error) {
	if len(key) != 64 {
		return nil, fmt.Errorf("invalid key size: must be %d bytes", 64)
	}

	chunkProcessor, err := processor.NewChunkProcessor(key, isEncryption)
	if err != nil {
		return nil, err
	}

	return &FileProcessor{
		chunkProcessor: chunkProcessor,
		workerCount:    runtime.NumCPU(),
	}, nil
}

// Process reads from reader, processes the data, and writes to writer
func (f *FileProcessor) Process(r io.Reader, w io.Writer, size int64) error {
	if err := f.validateInput(r, w); err != nil {
		return err
	}

	f.setProgressBar(size)
	return f.executePipeline(r, w)
}

// GetAesNonce returns the current AES nonce
func (f *FileProcessor) GetAesNonce() []byte {
	return f.chunkProcessor.AESCipher.GetNonce()
}

// GetChaCha20Nonce returns the current ChaCha20 nonce
func (f *FileProcessor) GetChaCha20Nonce() []byte {
	return f.chunkProcessor.ChaCha20Cipher.GetNonce()
}

// SetAesNonce sets a specific AES nonce
func (f *FileProcessor) SetAesNonce(aesNonce []byte) error {
	return f.chunkProcessor.AESCipher.SetNonce(aesNonce)
}

// SetChaCha20Nonce sets a specific ChaCha20 nonce
func (f *FileProcessor) SetChaCha20Nonce(chaCha20Nonce []byte) error {
	return f.chunkProcessor.ChaCha20Cipher.SetNonce(chaCha20Nonce)
}

func (f *FileProcessor) validateInput(r io.Reader, w io.Writer) error {
	if r == nil || w == nil {
		return fmt.Errorf("reader and writer must be non-nil")
	}
	return nil
}

func (f *FileProcessor) setProgressBar(size int64) {
	action := "Encrypting..."
	if !f.chunkProcessor.IsEncryption {
		action = "Decrypting..."
	}

	f.progressBar = progressbar.NewOptions64(
		size,
		progressbar.OptionSetDescription(action),
		progressbar.OptionUseANSICodes(false),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetTheme(progressbar.ThemeUnicode),
	)
}
