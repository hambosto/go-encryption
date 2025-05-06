package worker

import (
	"fmt"
	"io"
	"sync"
)

func (ws *WorkerStream) runPipeline(reader io.Reader, writer io.Writer) error {
	jobs := make(chan job, ws.workerCount)
	results := make(chan result, ws.workerCount)
	errChan := make(chan error, 1)

	// Start workers
	var workersWg sync.WaitGroup
	workersWg.Add(ws.workerCount)
	for range ws.workerCount {
		go func() {
			defer workersWg.Done()
			ws.processJobs(jobs, results)
		}()
	}

	// Start result writer goroutine
	var writerWg sync.WaitGroup
	writerWg.Add(1)
	go ws.writeResults(writer, results, &writerWg, errChan)

	// Read input and send jobs
	var readErr error
	if ws.processor.IsEncryption {
		readErr = ws.readEncryptChunks(reader, jobs)
	} else {
		readErr = ws.readDecryptChunks(reader, jobs)
	}

	// Close jobs channel to signal workers to exit
	close(jobs)

	// Wait for all workers to complete
	workersWg.Wait()

	// Close results channel to signal writer to exit
	close(results)

	// Wait for writer to complete
	writerWg.Wait()

	// Check for errors from the write goroutine
	select {
	case err := <-errChan:
		if readErr != nil {
			// If we have both read and write errors, combine them
			return fmt.Errorf("multiple errors: %v and %v", readErr, err)
		}
		return err
	default:
		return readErr
	}
}

func (ws *WorkerStream) processJobs(jobs <-chan job, results chan<- result) {
	for j := range jobs {
		output, err := ws.processor.ProcessChunk(j.data)
		size := len(j.data)
		if !ws.processor.IsEncryption {
			size = len(output)
		}
		results <- result{
			index: j.index,
			data:  output,
			size:  size,
			err:   err,
		}
	}
}
