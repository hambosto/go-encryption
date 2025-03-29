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
	for range f.workerCount {
		wg.Add(1)
		go f.processJobs(jobs, results, wg)
	}
}

func (f *FileProcessor) processJobs(jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

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
