package worker

import (
	"fmt"
	"io"
	"sync"
)

func (f *FileProcessor) executePipeline(r io.Reader, w io.Writer) error {
	jobs := make(chan Job, f.workerCount)
	results := make(chan Result, f.workerCount)
	errChan := make(chan error, 1)
	done := make(chan struct{}, f.workerCount)

	// Start the worker goroutines
	for range f.workerCount {
		go f.processJobs(jobs, results, done)
	}

	// Start the result collector
	var wg sync.WaitGroup
	wg.Add(1)
	go f.collectResults(w, results, &wg, errChan)

	// Distribute the jobs
	if err := f.distributeJobs(r, jobs, errChan); err != nil {
		return fmt.Errorf("job distribution failed: %w", err)
	}

	// Close the jobs channel to signal workers to exit
	close(jobs)

	// Wait for all workers to finish
	for i := 0; i < f.workerCount; i++ {
		<-done
	}

	// Close the results channel
	close(results)

	// Wait for result collector to finish
	wg.Wait()

	// Check for any errors
	select {
	case err := <-errChan:
		return err
	default:
		return nil
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
