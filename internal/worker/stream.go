package worker

import (
	"fmt"
	"io"
	"runtime"

	"github.com/hambosto/go-encryption/internal/processor"
	"github.com/schollz/progressbar/v3"
)

const (
	chunkSize      = 1024 * 1024
	defaultWorkers = 0
)

type WorkerStream struct {
	processor   *processor.ChunkProcessor
	progress    *progressbar.ProgressBar
	workerCount int
}

func NewWorkerStream(key []byte, encrypt bool) (*WorkerStream, error) {
	if len(key) != 64 {
		return nil, fmt.Errorf("key must be 64 bytes")
	}

	p, err := processor.NewChunkProcessor(key, encrypt)
	if err != nil {
		return nil, fmt.Errorf("failed to create chunk processor: %w", err)
	}

	workerCount := max(runtime.NumCPU(), 1)

	return &WorkerStream{
		processor:   p,
		workerCount: workerCount,
	}, nil
}

func (ws *WorkerStream) WithWorkerCount(count int) *WorkerStream {
	if count > 0 {
		ws.workerCount = count
	}
	return ws
}

func (ws *WorkerStream) Process(input io.Reader, output io.Writer, totalSize int64) error {
	if input == nil || output == nil {
		return fmt.Errorf("input and output streams must not be nil")
	}

	ws.initProgress(totalSize)
	return ws.runPipeline(input, output)
}

func (ws *WorkerStream) GetAESNonce() []byte {
	return ws.processor.AESCipher.GetNonce()
}

func (ws *WorkerStream) GetChaCha20Nonce() []byte {
	return ws.processor.ChaCha20Cipher.GetNonce()
}

func (ws *WorkerStream) SetAESNonce(nonce []byte) error {
	return ws.processor.AESCipher.SetNonce(nonce)
}

func (ws *WorkerStream) SetChaCha20Nonce(nonce []byte) error {
	return ws.processor.ChaCha20Cipher.SetNonce(nonce)
}

func (ws *WorkerStream) initProgress(size int64) {
	label := "Encrypting..."
	if !ws.processor.IsEncryption {
		label = "Decrypting..."
	}

	ws.progress = progressbar.NewOptions64(
		size,
		progressbar.OptionSetDescription(label),
		progressbar.OptionUseANSICodes(false),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetTheme(progressbar.ThemeUnicode),
	)
}
