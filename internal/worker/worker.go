package worker

import (
	"fmt"
	"io"
	"runtime"

	"github.com/hambosto/go-encryption/internal/processor"
	"github.com/schollz/progressbar/v3"
)

const ChunkSize = 1024 * 1024 // 1MB

type FileProcessor struct {
	chunkProcessor *processor.ChunkProcessor
	progressBar    *progressbar.ProgressBar
	workerCount    int
}

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

func (f *FileProcessor) Process(r io.Reader, w io.Writer, size int64) error {
	if err := f.validateInput(r, w); err != nil {
		return err
	}

	f.setProgressBar(size)
	return f.executePipeline(r, w)
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
