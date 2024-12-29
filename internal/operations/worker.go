package operations

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"github.com/schollz/progressbar/v3"
)

type Jobs struct {
	chunk []byte
	index uint32
}

type Results struct {
	index uint32
	data  []byte
	size  int
	err   error
}

func (f *FileProcessor) Process(r io.Reader, w io.Writer, size int64) error {
	if r == nil || w == nil {
		return fmt.Errorf("reader and writer must be non-nil")
	}

	action := "Encrypting..."
	if !f.chunkProcessor.isEncryption {
		action = "Decrypting..."
	}

	f.bar = progressbar.DefaultBytes(size, action)

	jobs := make(chan Jobs, f.workers)
	results := make(chan Results, f.workers)
	errChan := make(chan error, 1)

	return f.runProcessingPipeline(r, w, jobs, results, errChan)
}

func (f *FileProcessor) runProcessingPipeline(
	r io.Reader,
	w io.Writer,
	jobs chan Jobs,
	results chan Results,
	errChan chan error,
) error {
	var worker sync.WaitGroup
	var writer sync.WaitGroup

	for i := 0; i < f.workers; i++ {
		worker.Add(1)
		go f.worker(jobs, results, &worker)
	}

	writer.Add(1)
	go f.resultCollector(w, results, &writer, errChan)

	if err := f.distributeJobs(r, jobs, errChan); err != nil {
		return err
	}

	close(jobs)
	worker.Wait()
	close(results)
	writer.Wait()

	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}

func (f *FileProcessor) distributeJobs(
	r io.Reader,
	jobs chan<- Jobs,
	errChan chan error,
) error {
	switch f.chunkProcessor.isEncryption {
	case true:
		buffer := make([]byte, MaxChunkSize)

		var chunkIndex uint32
		for {
			n, err := r.Read(buffer)
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("failed to read chunk: %w", err)
			}
			chunk := make([]byte, n)
			copy(chunk, buffer[:n])

			select {
			case jobs <- Jobs{chunk: chunk, index: chunkIndex}:
				chunkIndex++
			case err := <-errChan:
				return fmt.Errorf("failed to enqueue chunk: %w", err)
			}
		}
	case false:
		buffer := make([]byte, 4)

		var chunkIndex uint32

		for {
			_, err := io.ReadFull(r, buffer)
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("failed to read chunk size: %w", err)
			}

			chunkSize := binary.BigEndian.Uint32(buffer)
			chunk := make([]byte, chunkSize)
			if _, err := io.ReadFull(r, chunk); err != nil {
				return fmt.Errorf("failed to read chunk data: %w", err)
			}

			select {
			case jobs <- Jobs{chunk: chunk, index: chunkIndex}:
				chunkIndex++
			case err := <-errChan:
				return err
			}
		}
	}

	return nil
}

func (f *FileProcessor) worker(
	jobs <-chan Jobs,
	results chan<- Results,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	for job := range jobs {
		processed, err := f.chunkProcessor.ProcessChunk(job.chunk)
		if f.chunkProcessor.isEncryption {
			results <- Results{
				index: job.index,
				data:  processed,
				size:  len(job.chunk),
				err:   err,
			}
		} else {
			results <- Results{
				index: job.index,
				data:  processed,
				size:  len(processed),
				err:   err,
			}
		}

	}
}

func (f *FileProcessor) resultCollector(
	w io.Writer,
	results <-chan Results,
	wg *sync.WaitGroup,
	errChan chan<- error,
) {
	defer wg.Done()
	pendingResults := make(map[uint32]Results)
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

func (f *FileProcessor) processOrderedResults(
	w io.Writer,
	pendingResults map[uint32]Results,
	nextIndex *uint32,
	errChan chan<- error,
) {
	for {
		chunk, exists := pendingResults[*nextIndex]
		if !exists {
			break
		}

		if f.chunkProcessor.isEncryption {

			sizeBuffer := make([]byte, 4)
			binary.BigEndian.PutUint32(sizeBuffer, uint32(len(chunk.data)))

			if _, err := w.Write(sizeBuffer); err != nil {
				errChan <- fmt.Errorf("failed to write chunk size: %w", err)
				return
			}
		}

		if _, err := w.Write(chunk.data); err != nil {
			errChan <- fmt.Errorf("failed to write chunk data: %w", err)
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
