package operations

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"github.com/schollz/progressbar/v3"
)

type Job struct {
	Data  []byte
	Index uint32
}

type Result struct {
	Index uint32
	Data  []byte
	Size  int
	Error error
}

func (f *FileProcessor) Process(r io.Reader, w io.Writer, size int64) error {
	if err := f.validateInputs(r, w); err != nil {
		return err
	}

	f.initializeProgressBar(size)
	return f.executePipeline(r, w)
}

func (f *FileProcessor) validateInputs(r io.Reader, w io.Writer) error {
	if r == nil || w == nil {
		return fmt.Errorf("reader and writer must be non-nil")
	}
	return nil
}

func (f *FileProcessor) initializeProgressBar(size int64) {
	action := "Encrypting..."
	if !f.chunkProcessor.isEncryption {
		action = "Decrypting..."
	}
	f.bar = progressbar.DefaultBytes(size, action)
}

func (f *FileProcessor) executePipeline(r io.Reader, w io.Writer) error {
	jobs := make(chan Job, f.workers)
	results := make(chan Result, f.workers)
	errChan := make(chan error, 1)

	var (
		workerGroup sync.WaitGroup
		writerGroup sync.WaitGroup
	)

	f.startWorkers(&workerGroup, jobs, results)

	writerGroup.Add(1)
	go f.collectResults(w, results, &writerGroup, errChan)

	if err := f.distributeJobs(r, jobs, errChan); err != nil {
		return fmt.Errorf("job distribution failed: %w", err)
	}

	return f.waitForCompletion(jobs, results, &workerGroup, &writerGroup, errChan)
}

func (f *FileProcessor) startWorkers(wg *sync.WaitGroup, jobs <-chan Job, results chan<- Result) {
	for i := 0; i < f.workers; i++ {
		wg.Add(1)
		go f.processJobs(jobs, results, wg)
	}
}

func (f *FileProcessor) processJobs(jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		processed, err := f.chunkProcessor.ProcessChunk(job.Data)
		size := len(processed)
		if f.chunkProcessor.isEncryption {
			size = len(job.Data)
		}

		results <- Result{
			Index: job.Index,
			Data:  processed,
			Size:  size,
			Error: err,
		}
	}
}

func (f *FileProcessor) distributeJobs(r io.Reader, jobs chan<- Job, errChan chan error) error {
	if f.chunkProcessor.isEncryption {
		return f.distributeEncryptionJobs(r, jobs, errChan)
	}
	return f.distributeDecryptionJobs(r, jobs, errChan)
}

func (f *FileProcessor) distributeEncryptionJobs(r io.Reader, jobs chan<- Job, errChan chan error) error {
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

		if err := f.sendJob(jobs, Job{Data: chunk, Index: chunkIndex}, errChan); err != nil {
			return err
		}
		chunkIndex++
	}
	return nil
}

func (f *FileProcessor) distributeDecryptionJobs(r io.Reader, jobs chan<- Job, errChan chan error) error {
	var chunkIndex uint32
	sizeBuffer := make([]byte, 4)

	for {
		if _, err := io.ReadFull(r, sizeBuffer); err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("failed to read chunk size: %w", err)
		}

		chunkSize := binary.BigEndian.Uint32(sizeBuffer)
		chunk := make([]byte, chunkSize)

		if _, err := io.ReadFull(r, chunk); err != nil {
			return fmt.Errorf("failed to read chunk data: %w", err)
		}

		if err := f.sendJob(jobs, Job{Data: chunk, Index: chunkIndex}, errChan); err != nil {
			return err
		}
		chunkIndex++
	}
	return nil
}

func (f *FileProcessor) sendJob(jobs chan<- Job, job Job, errChan chan error) error {
	select {
	case jobs <- job:
		return nil
	case err := <-errChan:
		return fmt.Errorf("failed to enqueue chunk: %w", err)
	}
}

func (f *FileProcessor) collectResults(w io.Writer, results <-chan Result, wg *sync.WaitGroup, errChan chan<- error) {
	defer wg.Done()

	pendingResults := make(map[uint32]Result)
	var nextIndex uint32

	for result := range results {
		if result.Error != nil {
			errChan <- fmt.Errorf("chunk %d processing failed: %w", result.Index, result.Error)
			return
		}

		pendingResults[result.Index] = result
		if err := f.writeOrderedResults(w, pendingResults, &nextIndex); err != nil {
			errChan <- err
			return
		}
	}
}

func (f *FileProcessor) writeOrderedResults(w io.Writer, pendingResults map[uint32]Result, nextIndex *uint32) error {
	for {
		result, exists := pendingResults[*nextIndex]
		if !exists {
			break
		}

		if err := f.writeResult(w, result); err != nil {
			return err
		}

		delete(pendingResults, *nextIndex)
		*nextIndex++
	}
	return nil
}

func (f *FileProcessor) writeResult(w io.Writer, result Result) error {
	if f.chunkProcessor.isEncryption {
		if err := f.writeChunkSize(w, len(result.Data)); err != nil {
			return err
		}
	}

	if _, err := w.Write(result.Data); err != nil {
		return fmt.Errorf("failed to write chunk data: %w", err)
	}

	if err := f.bar.Add(result.Size); err != nil {
		return fmt.Errorf("failed to update progress bar: %w", err)
	}

	return nil
}

func (f *FileProcessor) writeChunkSize(w io.Writer, size int) error {
	sizeBuffer := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBuffer, uint32(size))

	if _, err := w.Write(sizeBuffer); err != nil {
		return fmt.Errorf("failed to write chunk size: %w", err)
	}
	return nil
}

func (f *FileProcessor) waitForCompletion(
	jobs chan Job,
	results chan Result,
	workerGroup *sync.WaitGroup,
	writerGroup *sync.WaitGroup,
	errChan chan error,
) error {
	close(jobs)
	workerGroup.Wait()
	close(results)
	writerGroup.Wait()

	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}
