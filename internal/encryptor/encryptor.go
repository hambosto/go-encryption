package encryptor

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"runtime"
	"sync"

	"github.com/hambosto/go-encryption/internal/algorithms"
	"github.com/hambosto/go-encryption/internal/config"
	"github.com/hambosto/go-encryption/internal/encoding"
	"github.com/schollz/progressbar/v3"
)

type ChunkResult struct {
	index uint32
	data  []byte
	size  int
	err   error
}

type ChunkProcessor struct {
	serpentCipher  *algorithms.SerpentCipher
	chaCha20Cipher *algorithms.ChaCha20Cipher
	rsEncoder      *encoding.ReedSolomon
	bufferPool     sync.Pool
	compressPool   sync.Pool
}

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

func NewChunkProcessor(key []byte) (*ChunkProcessor, error) {
	serpentCipher, err := algorithms.NewSerpentCipher(key[:32])
	if err != nil {
		return nil, fmt.Errorf("failed to create serpent cipher: %w", err)
	}

	chaCha20Cipher, err := algorithms.NewChaCha20Cipher(key[32:])
	if err != nil {
		return nil, fmt.Errorf("failed to create chacha20 cipher: %w", err)
	}

	rsEncoder, err := encoding.NewReedSolomon(config.DataShards, config.ParityShards)
	if err != nil {
		return nil, fmt.Errorf("failed to create reed-solomon encoder: %w", err)
	}

	return &ChunkProcessor{
		serpentCipher:  serpentCipher,
		chaCha20Cipher: chaCha20Cipher,
		rsEncoder:      rsEncoder,
		bufferPool: sync.Pool{
			New: func() interface{} {
				buffer := make([]byte, config.MaxChunkSize)
				return &buffer
			},
		},
		compressPool: sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}, nil
}

func (cp *ChunkProcessor) processChunk(chunk []byte) ([]byte, error) {
	compressedData, err := cp.compressData(chunk)
	if err != nil {
		return nil, fmt.Errorf("failed to compress data: %w", err)
	}

	paddedData := cp.padData(compressedData)

	serpentEncrypted, err := cp.serpentCipher.Encrypt(paddedData)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	chaCha20Encrypted, err := cp.chaCha20Cipher.Encrypt(serpentEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	rsEncoded, err := cp.rsEncoder.Encode(chaCha20Encrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to encode data: %w", err)
	}

	return rsEncoded, nil
}

func (cp *ChunkProcessor) compressData(data []byte) ([]byte, error) {
	buffer := cp.compressPool.Get().(*bytes.Buffer)
	buffer.Reset()
	defer cp.compressPool.Put(buffer)

	zw, err := zlib.NewWriterLevel(buffer, zlib.BestSpeed)
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib writer: %w", err)
	}

	if _, err := zw.Write(data); err != nil {
		zw.Close()
		return nil, fmt.Errorf("failed to write data to zlib writer: %w", err)
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zlib writer: %w", err)
	}

	result := make([]byte, buffer.Len())
	copy(result, buffer.Bytes())

	return result, nil
}

func (cp *ChunkProcessor) padData(data []byte) []byte {
	sizeHeader := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeHeader, uint32(len(data)))

	alignedSize := (len(data) + 15) & ^15
	if alignedSize > len(data) {
		padding := make([]byte, alignedSize-len(data))
		data = append(data, padding...)
	}

	return append(sizeHeader, data...)
}

func (f *FileEncryptor) Encrypt(r io.Reader, w io.Writer, size int64) error {
	if r == nil {
		return fmt.Errorf("invalid reader: must be non-nil")
	}

	if w == nil {
		return fmt.Errorf("invalid writer: must be non-nil")
	}

	f.bar = progressbar.DefaultBytes(size, "Encrypting...")

	jobs := make(chan struct {
		chunk []byte
		index uint32
	}, f.workers)

	results := make(chan ChunkResult, f.workers)
	errChan := make(chan error, 1)

	var wg sync.WaitGroup
	for i := 0; i < f.workers; i++ {
		wg.Add(1)
		go f.encryptWorker(jobs, results, &wg)
	}

	var writeWg sync.WaitGroup
	writeWg.Add(1)
	go f.resultCollector(w, results, &writeWg, errChan)

	buffer := f.chunkProcessor.bufferPool.Get().(*[]byte)
	defer f.chunkProcessor.bufferPool.Put(buffer)

	var chunkIndex uint32
	for {
		n, err := r.Read(*buffer)

		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("failed to read chunk: %w", err)
		}

		chunk := make([]byte, n)
		copy(chunk, (*buffer)[:n])

		select {
		case jobs <- struct {
			chunk []byte
			index uint32
		}{chunk, chunkIndex}:
		case err := <-errChan:
			close(jobs)
			return fmt.Errorf("failed to enqueue chunk: %w", err)
		}
		chunkIndex++
	}

	close(jobs)
	wg.Wait()
	close(results)
	writeWg.Wait()

	select {
	case err := <-errChan:
		return fmt.Errorf("failed to write chunks: %w", err)
	default:
		return nil
	}
}

func (f *FileEncryptor) GetNonce() ([]byte, []byte) {
	return f.chunkProcessor.serpentCipher.GetNonce(), f.chunkProcessor.chaCha20Cipher.GetNonce()
}

func (f *FileEncryptor) writeChunk(w io.Writer, chunk []byte) error {
	sizeBuffer := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBuffer, uint32(len(chunk)))

	if _, err := w.Write(sizeBuffer); err != nil {
		return fmt.Errorf("failed to write chunk size: %w", err)
	}

	if _, err := w.Write(chunk); err != nil {
		return fmt.Errorf("failed to write chunk data: %w", err)
	}

	return nil
}

func (f *FileEncryptor) encryptWorker(jobs <-chan struct {
	chunk []byte
	index uint32
}, results chan<- ChunkResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		processed, err := f.chunkProcessor.processChunk(job.chunk)
		results <- ChunkResult{
			index: job.index,
			data:  processed,
			size:  len(job.chunk),
			err:   err,
		}
	}
}

func (f *FileEncryptor) resultCollector(w io.Writer, results <-chan ChunkResult, wg *sync.WaitGroup, errChan chan<- error) {
	defer wg.Done()

	pendingResults := make(map[uint32]ChunkResult)
	nextIndex := uint32(0)

	for result := range results {
		if result.err != nil {
			errChan <- fmt.Errorf("failed to process chunk %d: %w", result.index, result.err)
			return
		}

		pendingResults[result.index] = result

		for {
			if chunk, ok := pendingResults[nextIndex]; ok {
				if err := f.writeChunk(w, chunk.data); err != nil {
					errChan <- fmt.Errorf("failed to write chunk %d: %w", chunk.index, err)
					return
				}

				if err := f.bar.Add(chunk.size); err != nil {
					errChan <- fmt.Errorf("failed to update progress bar: %w", err)
					return
				}

				delete(pendingResults, nextIndex)
				nextIndex++
			} else {
				break
			}
		}
	}
}
