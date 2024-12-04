package encryptor

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"github.com/schollz/progressbar/v3"
)

type EncryptJob struct {
	chunk []byte
	index uint32
}

type EncryptResult struct {
	index uint32
	data  []byte
	size  int
	err   error
}

func (f *FileEncryptor) Encrypt(r io.Reader, w io.Writer, size int64) error {
	if r == nil || w == nil {
		return fmt.Errorf("reader and writer must be non-nil")
	}

	f.bar = progressbar.DefaultBytes(size, "Encrypting...")

	jobs := make(chan EncryptJob, f.workers)
	results := make(chan EncryptResult, f.workers)
	errChan := make(chan error, 1)

	var wg sync.WaitGroup
	for i := 0; i < f.workers; i++ {
		wg.Add(1)
		go f.encryptWorker(jobs, results, &wg)
	}

	var writeWg sync.WaitGroup
	writeWg.Add(1)
	go f.resultCollector(w, results, &writeWg, errChan)

	if err := f.distributeJobs(r, jobs, errChan); err != nil {
		return err
	}

	close(jobs)
	wg.Wait()

	close(results)
	writeWg.Wait()

	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}

func (f *FileEncryptor) distributeJobs(r io.Reader, jobs chan<- EncryptJob, errChan chan error) error {
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
		case jobs <- EncryptJob{chunk: chunk, index: chunkIndex}:
			chunkIndex++
		case err := <-errChan:
			return fmt.Errorf("failed to enqueue chunk: %w", err)
		}
	}
	return nil
}

func (f *FileEncryptor) encryptWorker(jobs <-chan EncryptJob, results chan<- EncryptResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		processed, err := f.chunkProcessor.ProcessChunk(job.chunk)
		results <- EncryptResult{
			index: job.index,
			data:  processed,
			size:  len(job.chunk),
			err:   err,
		}
	}
}

func (f *FileEncryptor) resultCollector(w io.Writer, results <-chan EncryptResult, wg *sync.WaitGroup, errChan chan<- error) {
	defer wg.Done()

	pendingResults := make(map[uint32]EncryptResult)
	nextIndex := uint32(0)

	for result := range results {
		if result.err != nil {
			errChan <- fmt.Errorf("failed to process chunk %d: %w", result.index, result.err)
			return
		}

		pendingResults[result.index] = result

		for {
			if chunk, exists := pendingResults[nextIndex]; exists {
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
