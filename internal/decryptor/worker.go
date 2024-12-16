package decryptor

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync"

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

	return f.runDecryptionPipeline(r, w, jobs, results, errChan)
}

func (f *FileDecryptor) runDecryptionPipeline(
	r io.Reader,
	w io.Writer,
	jobs chan DecryptJob,
	results chan ChunkResult,
	errChan chan error,
) error {
	var workerWg sync.WaitGroup
	var writeWg sync.WaitGroup

	for i := 0; i < f.workers; i++ {
		workerWg.Add(1)
		go f.decryptWorker(jobs, results, &workerWg)
	}

	writeWg.Add(1)
	go f.resultCollector(w, results, &writeWg, errChan)

	if err := f.enqueueJobs(r, jobs, errChan); err != nil {
		return err
	}

	close(jobs)
	workerWg.Wait()
	close(results)
	writeWg.Wait()

	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}

func (f *FileDecryptor) enqueueJobs(
	r io.Reader,
	jobs chan<- DecryptJob,
	errChan chan error,
) error {
	sizeBuffer := make([]byte, 4)
	var chunkIndex uint32

	for {
		_, err := io.ReadFull(r, sizeBuffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read chunk size: %w", err)
		}

		chunkSize := binary.BigEndian.Uint32(sizeBuffer)
		chunk := make([]byte, chunkSize)
		if _, err := io.ReadFull(r, chunk); err != nil {
			return fmt.Errorf("failed to read chunk data: %w", err)
		}

		select {
		case jobs <- DecryptJob{data: chunk, index: chunkIndex}:
			chunkIndex++
		case err := <-errChan:
			return err
		}
	}
	return nil
}

func (f *FileDecryptor) decryptWorker(
	jobs <-chan DecryptJob,
	results chan<- ChunkResult,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	for job := range jobs {
		processed, err := f.chunkProcessor.ProcessChunk(job.data)
		if err != nil {
			results <- ChunkResult{index: job.index, err: err}
			continue
		}

		decompressed, err := f.chunkProcessor.DecompressData(processed)
		results <- ChunkResult{
			index: job.index,
			data:  decompressed,
			size:  len(decompressed),
			err:   err,
		}
	}
}

func (f *FileDecryptor) resultCollector(
	w io.Writer,
	results <-chan ChunkResult,
	wg *sync.WaitGroup,
	errChan chan<- error,
) {
	defer wg.Done()
	pendingResults := make(map[uint32]ChunkResult)
	nextIndex := uint32(0)

	for result := range results {
		if result.err != nil {
			errChan <- fmt.Errorf("failed to process chunk %d: %w", result.index, result.err)
			return
		}

		pendingResults[result.index] = result
		f.processOrderedResults(w, pendingResults, &nextIndex, errChan)
	}
}

func (f *FileDecryptor) processOrderedResults(
	w io.Writer,
	pendingResults map[uint32]ChunkResult,
	nextIndex *uint32,
	errChan chan<- error,
) {
	for {
		chunk, exists := pendingResults[*nextIndex]
		if !exists {
			break
		}

		if err := f.writeChunk(w, chunk.data); err != nil {
			errChan <- fmt.Errorf("failed to write chunk %d: %w", chunk.index, err)
			return
		}

		if err := f.bar.Add(chunk.size); err != nil {
			errChan <- fmt.Errorf("failed to update progress bar: %w", err)
			return
		}

		delete(pendingResults, *nextIndex)
		*nextIndex++
	}
}

func (f *FileDecryptor) writeChunk(w io.Writer, chunk []byte) error {
	if _, err := w.Write(chunk); err != nil {
		return fmt.Errorf("failed to write chunk data: %w", err)
	}
	return nil
}
