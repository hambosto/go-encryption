package worker

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync"
)

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
