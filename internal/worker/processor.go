package worker

import (
	"encoding/binary"
	"fmt"
	"io"
)

func (f *FileProcessor) distributeJobs(r io.Reader, jobs chan<- Job, errChan chan error) error {
	if f.chunkProcessor.IsEncryption {
		return f.distributeEncryptionJobs(r, jobs, errChan)
	}
	return f.distributeDecryptionJobs(r, jobs, errChan)
}

func (f *FileProcessor) distributeEncryptionJobs(r io.Reader, jobs chan<- Job, errChan chan error) error {
	buffer := make([]byte, ChunkSize)
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

func (f *FileProcessor) processJobs(jobs <-chan Job, results chan<- Result, done chan<- struct{}) {
	for job := range jobs {
		processed, err := f.chunkProcessor.ProcessChunk(job.Data)
		size := len(processed)
		if f.chunkProcessor.IsEncryption {
			size = len(job.Data)
		}

		results <- Result{
			Index: job.Index,
			Data:  processed,
			Size:  size,
			Error: err,
		}
	}
	done <- struct{}{}
}

func (f *FileProcessor) writeResult(w io.Writer, result Result) error {
	if f.chunkProcessor.IsEncryption {
		if err := f.writeChunkSize(w, len(result.Data)); err != nil {
			return err
		}
	}

	if _, err := w.Write(result.Data); err != nil {
		return fmt.Errorf("failed to write chunk data: %w", err)
	}

	if err := f.progressBar.Add(result.Size); err != nil {
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
