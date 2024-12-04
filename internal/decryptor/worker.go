package decryptor

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"github.com/hambosto/go-encryption/internal/constants"
	"github.com/schollz/progressbar/v3"
)

type DecryptJob struct {
	data  []byte
	index uint32
}

type ChunkResult struct {
	index uint32
	data  []byte
	size  int
	err   error
}

func (f *FileDecryptor) Decrypt(r io.Reader, w io.Writer, size int64) error {
	if r == nil || w == nil {
		return fmt.Errorf("reader and writer must be non-nil")
	}

	f.bar = progressbar.DefaultBytes(size, "Decrypting...")

	jobs := make(chan DecryptJob, f.workers)
	results := make(chan ChunkResult, f.workers)
	errChan := make(chan error, 1)

	// Launch workers.
	var wg sync.WaitGroup
	for i := 0; i < f.workers; i++ {
		wg.Add(1)
		go f.decryptWorker(jobs, results, &wg)
	}

	// Launch result collector.
	var writeWg sync.WaitGroup
	writeWg.Add(1)
	go f.resultCollector(w, results, &writeWg, errChan)

	// Parse chunks and enqueue jobs.
	if err := f.enqueueJobs(r, jobs, errChan); err != nil {
		return err
	}

	// Wait for workers and collector to finish.
	close(jobs)
	wg.Wait()
	close(results)
	writeWg.Wait()

	// Check for any errors during result collection.
	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}

// enqueueJobs parses chunks from the reader and sends them to the jobs channel.
func (f *FileDecryptor) enqueueJobs(r io.Reader, jobs chan<- DecryptJob, errChan chan error) error {
	sizeBuffer := make([]byte, 4)
	var chunkIndex uint32

	for {
		// Read the size of the next chunk.
		_, err := io.ReadFull(r, sizeBuffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read chunk size: %w", err)
		}

		// Validate chunk size.
		chunkSize := binary.BigEndian.Uint32(sizeBuffer)
		if err := f.validateChunkSize(chunkSize); err != nil {
			return err
		}

		// Read the actual chunk data.
		chunk := make([]byte, chunkSize)
		if _, err := io.ReadFull(r, chunk); err != nil {
			return fmt.Errorf("failed to read chunk data: %w", err)
		}

		// Send job to the workers or handle errors.
		select {
		case jobs <- DecryptJob{data: chunk, index: chunkIndex}:
			chunkIndex++
		case err := <-errChan:
			return err
		}
	}
	return nil
}

// validateChunkSize ensures the chunk size is valid.
func (f *FileDecryptor) validateChunkSize(chunkSize uint32) error {
	if chunkSize == 0 || chunkSize > constants.MaxEncryptedChunkSize {
		return fmt.Errorf("invalid chunk size: must be between 1 and %d", constants.MaxEncryptedChunkSize)
	}
	if chunkSize%(constants.DataShards+constants.ParityShards) != 0 {
		return fmt.Errorf("invalid chunk size: must be a multiple of %d", constants.DataShards+constants.ParityShards)
	}
	return nil
}

// decryptWorker processes jobs and sends results to the results channel.
func (f *FileDecryptor) decryptWorker(jobs <-chan DecryptJob, results chan<- ChunkResult, wg *sync.WaitGroup) {
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

// resultCollector writes processed chunks to the writer in the correct order.
func (f *FileDecryptor) resultCollector(w io.Writer, results <-chan ChunkResult, wg *sync.WaitGroup, errChan chan<- error) {
	defer wg.Done()

	pendingResults := make(map[uint32]ChunkResult)
	nextIndex := uint32(0)

	for result := range results {
		if result.err != nil {
			errChan <- fmt.Errorf("failed to process chunk %d: %w", result.index, result.err)
			return
		}

		// Add result to pending map.
		pendingResults[result.index] = result

		// Write chunks in order if available.
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

// writeChunk writes a single chunk to the writer.
func (f *FileDecryptor) writeChunk(w io.Writer, chunk []byte) error {
	if _, err := w.Write(chunk); err != nil {
		return fmt.Errorf("failed to write chunk data: %w", err)
	}
	return nil
}
