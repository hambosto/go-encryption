package encryptor

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"github.com/schollz/progressbar/v3"
)

type ChunkResult struct {
	index uint32
	data  []byte
	size  int
	err   error
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
}, results chan<- ChunkResult, wg *sync.WaitGroup,
) {
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
