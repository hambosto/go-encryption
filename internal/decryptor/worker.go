package decryptor

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"github.com/hambosto/go-encryption/internal/constants"
	"github.com/schollz/progressbar/v3"
)

type ChunkResult struct {
	index uint32
	data  []byte
	size  int
	err   error
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

		if chunkSize == 0 || chunkSize > constants.MaxEncryptedChunkSize {
			close(jobs)
			return fmt.Errorf("invalid chunk size: must be between 0 and %d", constants.MaxEncryptedChunkSize)
		}

		if chunkSize%(constants.DataShards+constants.ParityShards) != 0 {
			close(jobs)
			return fmt.Errorf("invalid chunk size: must be a multiple of %d", constants.DataShards+constants.ParityShards)
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

func (f *FileDecryptor) decryptWorker(jobs <-chan struct {
	data  []byte
	index uint32
}, results chan<- ChunkResult, wg *sync.WaitGroup,
) {
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

func (f *FileDecryptor) writeChunk(w io.Writer, chunk []byte) error {
	if _, err := w.Write(chunk); err != nil {
		return fmt.Errorf("failed to write chunk data: %w", err)
	}
	return nil
}
