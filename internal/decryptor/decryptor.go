package decryptor

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"runtime"
	"sync"

	"github.com/hambosto/go-encryption/pkg/crypto/algorithms"
	"github.com/hambosto/go-encryption/pkg/crypto/config"
	"github.com/hambosto/go-encryption/pkg/crypto/encoding"
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
	rsDecoder      *encoding.ReedSolomon
	bufferPool     sync.Pool
	decompressPool sync.Pool
}

type FileDecryptor struct {
	chunkProcessor *ChunkProcessor
	bar            *progressbar.ProgressBar
	workers        int
}

func NewFileDecryptor(key []byte) (*FileDecryptor, error) {
	if len(key) < 64 {
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

func NewChunkProcessor(key []byte) (*ChunkProcessor, error) {
	serpentCipher, err := algorithms.NewSerpentCipher(key[:32])
	if err != nil {
		return nil, fmt.Errorf("failed to create serpent cipher: %w", err)
	}

	chaCha20Cipher, err := algorithms.NewChaCha20Cipher(key[32:])
	if err != nil {
		return nil, fmt.Errorf("failed to create chacha20 cipher: %w", err)
	}

	rsDecoder, err := encoding.NewReedSolomon(config.DataShards, config.ParityShards)
	if err != nil {
		return nil, fmt.Errorf("failed to create reed solomon decoder: %w", err)
	}

	return &ChunkProcessor{
		serpentCipher:  serpentCipher,
		chaCha20Cipher: chaCha20Cipher,
		rsDecoder:      rsDecoder,
		bufferPool: sync.Pool{
			New: func() interface{} {
				buffer := make([]byte, config.MaxEncryptedChunkSize)
				return &buffer
			},
		},
		decompressPool: sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}, nil
}

func (cp *ChunkProcessor) processChunk(chunk []byte) ([]byte, error) {
	decodedData, err := cp.rsDecoder.Decode(chunk)
	if err != nil {
		return nil, fmt.Errorf("failed to decode data: %w", err)
	}

	chaCha20Decrypted, err := cp.chaCha20Cipher.Decrypt(decodedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	serpentDecrypted, err := cp.serpentCipher.Decrypt(chaCha20Decrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return serpentDecrypted, nil
}

func (cp *ChunkProcessor) decompressData(data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("invalid data size: must be at least 4 bytes")
	}

	compressedSize := binary.BigEndian.Uint32(data[:4])
	if compressedSize > uint32(len(data)-4) {
		return nil, fmt.Errorf("invalid data size: must be at least %d bytes", compressedSize)
	}

	compressedData := data[4 : 4+compressedSize]

	buffer := cp.decompressPool.Get().(*bytes.Buffer)
	buffer.Reset()
	defer cp.decompressPool.Put(buffer)

	zr, err := zlib.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader: %w", err)
	}
	defer zr.Close()

	if _, err := io.Copy(buffer, zr); err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}

	result := make([]byte, buffer.Len())
	copy(result, buffer.Bytes())

	return result, nil
}

func (f *FileDecryptor) Decrypt(r io.Reader, w io.Writer, size int64) error {
	if r == nil {
		return fmt.Errorf("invalid reader: must be non-nil")
	}

	if w == nil {
		return fmt.Errorf("invalid writer: must be non-nil")
	}

	f.bar = progressbar.DefaultBytes(size, "Decrypting...")

	jobs := make(chan struct {
		data  []byte
		index uint32
	}, f.workers)

	results := make(chan ChunkResult, f.workers)
	errChan := make(chan error, 1)

	var wg sync.WaitGroup
	for i := 0; i < f.workers; i++ {
		wg.Add(1)
		go f.decryptWorker(jobs, results, &wg)
	}

	var writeWg sync.WaitGroup
	writeWg.Add(1)
	go f.resultCollector(w, results, &writeWg, errChan)

	sizeBuffer := make([]byte, 4)
	var chunkIndex uint32
	for {
		_, err := r.Read(sizeBuffer)
		if err == io.EOF {
			break
		}

		if err != nil {
			close(jobs)
			return fmt.Errorf("failed to read chunk size: %w", err)
		}

		chunkSize := binary.BigEndian.Uint32(sizeBuffer)

		if chunkSize == 0 || chunkSize > config.MaxEncryptedChunkSize {
			close(jobs)
			return fmt.Errorf("invalid chunk size: must be between 0 and %d", config.MaxEncryptedChunkSize)
		}

		if chunkSize%(config.DataShards+config.ParityShards) != 0 {
			close(jobs)
			return fmt.Errorf("invalid chunk size: must be a multiple of %d", config.DataShards+config.ParityShards)
		}

		chunk := make([]byte, chunkSize)
		if _, err := io.ReadFull(r, chunk); err != nil {
			close(jobs)
			return fmt.Errorf("failed to read chunk data: %w", err)
		}

		select {
		case jobs <- struct {
			data  []byte
			index uint32
		}{chunk, chunkIndex}:
		case err := <-errChan:
			close(jobs)
			return err
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

func (f *FileDecryptor) SetNonce(serpentNonce, chacha20Nonce []byte) error {
	if err := f.chunkProcessor.serpentCipher.SetNonce(serpentNonce); err != nil {
		return fmt.Errorf("failed to set serpent nonce: %w", err)
	}

	if err := f.chunkProcessor.chaCha20Cipher.SetNonce(chacha20Nonce); err != nil {
		return fmt.Errorf("failed to set chacha20 nonce: %w", err)
	}

	return nil
}

func (f *FileDecryptor) writeChunk(w io.Writer, chunk []byte) error {
	if _, err := w.Write(chunk); err != nil {
		return fmt.Errorf("failed to write chunk data: %w", err)
	}
	return nil
}

func (f *FileDecryptor) decryptWorker(jobs <-chan struct {
	data  []byte
	index uint32
}, results chan<- ChunkResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		processed, err := f.chunkProcessor.processChunk(job.data)
		if err != nil {
			results <- ChunkResult{index: job.index, err: err}
			continue
		}

		decompressed, err := f.chunkProcessor.decompressData(processed)
		results <- ChunkResult{
			index: job.index,
			data:  decompressed,
			size:  len(decompressed),
			err:   err,
		}
	}
}

func (f *FileDecryptor) resultCollector(w io.Writer, results <-chan ChunkResult, wg *sync.WaitGroup, errChan chan<- error) {
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
